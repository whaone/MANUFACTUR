package sync

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"

	"manufactpro/backend/internal/db"
	"manufactpro/backend/internal/material"
	"manufactpro/backend/internal/product"
	"manufactpro/backend/internal/supplier"
)

// Item is one queued offline operation from the client.
type Item struct {
	ID              string          `json:"id"`
	Operation       string          `json:"operation"` // create | update | delete
	Entity          string          `json:"entity"`    // material | supplier
	Payload         json.RawMessage `json:"payload"`
	ClientTimestamp int64           `json:"client_timestamp"`
}

type Request struct {
	Items []Item `json:"items"`
}

// Result reports the outcome of one item so the client can reconcile its queue.
type Result struct {
	ID       string `json:"id"`     // client queue id (echoed back)
	Status   string `json:"status"` // synced | failed
	ServerID string `json:"server_id,omitempty"`
	Error    string `json:"error,omitempty"`
}

type Response struct {
	Results []Result `json:"results"`
	Synced  int      `json:"synced"`
	Failed  int      `json:"failed"`
}

// idHolder extracts the server id embedded in update/delete payloads.
type idHolder struct {
	ID string `json:"id"`
}

// Process applies each item independently. A single failure never aborts the batch.
// Idempotent: a successfully-applied op_id is cached in sync_processed, so replays
// (e.g. lost responses) return the same result without re-applying. Failed ops are
// NOT cached, so the client can retry them after fixing the cause.
func Process(ctx context.Context, workspaceID uuid.UUID, req Request) Response {
	resp := Response{Results: make([]Result, 0, len(req.Items))}
	for _, it := range req.Items {
		r := processOne(ctx, workspaceID, it)
		if r.Status == "synced" {
			resp.Synced++
		} else {
			resp.Failed++
		}
		resp.Results = append(resp.Results, r)
	}
	return resp
}

func processOne(ctx context.Context, workspaceID uuid.UUID, it Item) Result {
	if it.ID == "" {
		return Result{ID: it.ID, Status: "failed", Error: "missing op id"}
	}
	// Replay guard: return the cached result of a previously-synced op.
	if cached, ok := lookupProcessed(ctx, workspaceID, it.ID); ok {
		return cached
	}
	serverID, err := apply(ctx, workspaceID, it)
	r := Result{ID: it.ID}
	if err != nil {
		r.Status = "failed"
		r.Error = err.Error()
		return r // not recorded → client may retry
	}
	r.Status = "synced"
	r.ServerID = serverID
	recordProcessed(ctx, workspaceID, it.ID, r)
	return r
}

func lookupProcessed(ctx context.Context, workspaceID uuid.UUID, opID string) (Result, bool) {
	var r Result
	var serverID, errMsg *string
	err := db.Pool.QueryRow(ctx,
		`SELECT status, server_id, error FROM sync_processed WHERE workspace_id=$1 AND op_id=$2`,
		workspaceID, opID,
	).Scan(&r.Status, &serverID, &errMsg)
	if err != nil {
		return Result{}, false
	}
	r.ID = opID
	if serverID != nil {
		r.ServerID = *serverID
	}
	if errMsg != nil {
		r.Error = *errMsg
	}
	return r, true
}

func recordProcessed(ctx context.Context, workspaceID uuid.UUID, opID string, r Result) {
	// Best-effort: failure to record only risks a duplicate on replay, never data loss.
	_, _ = db.Pool.Exec(ctx,
		`INSERT INTO sync_processed (workspace_id, op_id, status, server_id, error)
		 VALUES ($1,$2,$3,NULLIF($4,''),NULLIF($5,''))
		 ON CONFLICT (workspace_id, op_id) DO NOTHING`,
		workspaceID, opID, r.Status, r.ServerID, r.Error,
	)
}

func apply(ctx context.Context, wsID uuid.UUID, it Item) (string, error) {
	switch it.Entity {
	case "material":
		return applyMaterial(ctx, wsID, it)
	case "supplier":
		return applySupplier(ctx, wsID, it)
	case "product":
		return applyProduct(ctx, wsID, it)
	default:
		return "", fmt.Errorf("unsupported entity: %s", it.Entity)
	}
}

func payloadID(p json.RawMessage) (uuid.UUID, error) {
	var h idHolder
	if err := json.Unmarshal(p, &h); err != nil {
		return uuid.Nil, fmt.Errorf("invalid payload: %w", err)
	}
	id, err := uuid.Parse(h.ID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("missing or invalid id in payload")
	}
	return id, nil
}

func applyMaterial(ctx context.Context, wsID uuid.UUID, it Item) (string, error) {
	switch it.Operation {
	case "create":
		var in material.CreateInput
		if err := json.Unmarshal(it.Payload, &in); err != nil {
			return "", fmt.Errorf("invalid payload: %w", err)
		}
		m, err := material.Create(ctx, wsID, in)
		if err != nil {
			return "", err
		}
		return m.ID.String(), nil
	case "update":
		id, err := payloadID(it.Payload)
		if err != nil {
			return "", err
		}
		var in material.UpdateInput
		if err := json.Unmarshal(it.Payload, &in); err != nil {
			return "", fmt.Errorf("invalid payload: %w", err)
		}
		if _, err := material.Update(ctx, wsID, id, in); err != nil {
			return "", err
		}
		return id.String(), nil
	case "delete":
		id, err := payloadID(it.Payload)
		if err != nil {
			return "", err
		}
		if err := material.Delete(ctx, wsID, id); err != nil {
			return "", err
		}
		return id.String(), nil
	default:
		return "", fmt.Errorf("unsupported operation: %s", it.Operation)
	}
}

func applyProduct(ctx context.Context, wsID uuid.UUID, it Item) (string, error) {
	switch it.Operation {
	case "create":
		var in product.CreateProductInput
		if err := json.Unmarshal(it.Payload, &in); err != nil {
			return "", fmt.Errorf("invalid payload: %w", err)
		}
		p, err := product.Create(ctx, wsID, in)
		if err != nil {
			return "", err
		}
		return p.ID.String(), nil
	case "update":
		id, err := payloadID(it.Payload)
		if err != nil {
			return "", err
		}
		var in product.UpdateProductInput
		if err := json.Unmarshal(it.Payload, &in); err != nil {
			return "", fmt.Errorf("invalid payload: %w", err)
		}
		if _, err := product.Update(ctx, wsID, id, in); err != nil {
			return "", err
		}
		return id.String(), nil
	case "delete":
		id, err := payloadID(it.Payload)
		if err != nil {
			return "", err
		}
		if err := product.Delete(ctx, wsID, id); err != nil {
			return "", err
		}
		return id.String(), nil
	default:
		return "", fmt.Errorf("unsupported operation: %s", it.Operation)
	}
}

func applySupplier(ctx context.Context, wsID uuid.UUID, it Item) (string, error) {
	switch it.Operation {
	case "create":
		var in supplier.CreateInput
		if err := json.Unmarshal(it.Payload, &in); err != nil {
			return "", fmt.Errorf("invalid payload: %w", err)
		}
		s, err := supplier.Create(ctx, wsID, in)
		if err != nil {
			return "", err
		}
		return s.ID.String(), nil
	case "update":
		id, err := payloadID(it.Payload)
		if err != nil {
			return "", err
		}
		var in supplier.UpdateInput
		if err := json.Unmarshal(it.Payload, &in); err != nil {
			return "", fmt.Errorf("invalid payload: %w", err)
		}
		if _, err := supplier.Update(ctx, wsID, id, in); err != nil {
			return "", err
		}
		return id.String(), nil
	case "delete":
		id, err := payloadID(it.Payload)
		if err != nil {
			return "", err
		}
		if err := supplier.Delete(ctx, wsID, id); err != nil {
			return "", err
		}
		return id.String(), nil
	default:
		return "", fmt.Errorf("unsupported operation: %s", it.Operation)
	}
}

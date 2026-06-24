package warehouse

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"manufactpro/backend/internal/db"
)

type Branch struct {
	ID          uuid.UUID `json:"id"`
	WorkspaceID uuid.UUID `json:"workspace_id"`
	Name        string    `json:"name"`
	Address     string    `json:"address"`
	CreatedAt   time.Time `json:"created_at"`
}

type Warehouse struct {
	ID        uuid.UUID `json:"id"`
	BranchID  uuid.UUID `json:"branch_id"`
	Name      string    `json:"name"`
	Code      string    `json:"code"`
	IsDefault bool      `json:"is_default"`
	CreatedAt time.Time `json:"created_at"`
}

type CreateBranchInput struct {
	Name    string `json:"name"`
	Address string `json:"address"`
}

type CreateWarehouseInput struct {
	BranchID  uuid.UUID `json:"branch_id"`
	Name      string    `json:"name"`
	Code      string    `json:"code"`
	IsDefault bool      `json:"is_default"`
}

func ListBranches(ctx context.Context, workspaceID uuid.UUID) ([]Branch, error) {
	rows, err := db.Pool.Query(ctx,
		`SELECT id, workspace_id, name, COALESCE(address,''), created_at
		 FROM branches WHERE workspace_id=$1 ORDER BY name`,
		workspaceID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []Branch
	for rows.Next() {
		var b Branch
		if err := rows.Scan(&b.ID, &b.WorkspaceID, &b.Name, &b.Address, &b.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, b)
	}
	if list == nil {
		list = []Branch{}
	}
	return list, nil
}

func CreateBranch(ctx context.Context, workspaceID uuid.UUID, in CreateBranchInput) (*Branch, error) {
	if in.Name == "" {
		return nil, fmt.Errorf("name required")
	}
	var b Branch
	err := db.Pool.QueryRow(ctx,
		`INSERT INTO branches(workspace_id,name,address) VALUES($1,$2,$3)
		 RETURNING id, workspace_id, name, COALESCE(address,''), created_at`,
		workspaceID, in.Name, in.Address,
	).Scan(&b.ID, &b.WorkspaceID, &b.Name, &b.Address, &b.CreatedAt)
	return &b, err
}

func ListWarehouses(ctx context.Context, workspaceID uuid.UUID) ([]Warehouse, error) {
	rows, err := db.Pool.Query(ctx,
		`SELECT w.id, w.branch_id, w.name, w.code, w.is_default, w.created_at
		 FROM warehouses w
		 JOIN branches b ON b.id = w.branch_id
		 WHERE b.workspace_id=$1 ORDER BY w.name`,
		workspaceID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []Warehouse
	for rows.Next() {
		var w Warehouse
		if err := rows.Scan(&w.ID, &w.BranchID, &w.Name, &w.Code, &w.IsDefault, &w.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, w)
	}
	if list == nil {
		list = []Warehouse{}
	}
	return list, nil
}

func CreateWarehouse(ctx context.Context, workspaceID uuid.UUID, in CreateWarehouseInput) (*Warehouse, error) {
	if in.Name == "" || in.Code == "" {
		return nil, fmt.Errorf("name and code required")
	}
	// verify branch belongs to workspace
	var count int
	err := db.Pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM branches WHERE id=$1 AND workspace_id=$2`,
		in.BranchID, workspaceID,
	).Scan(&count)
	if err != nil || count == 0 {
		return nil, fmt.Errorf("branch not found")
	}

	var w Warehouse
	err = db.Pool.QueryRow(ctx,
		`INSERT INTO warehouses(branch_id,name,code,is_default) VALUES($1,$2,$3,$4)
		 RETURNING id, branch_id, name, code, is_default, created_at`,
		in.BranchID, in.Name, in.Code, in.IsDefault,
	).Scan(&w.ID, &w.BranchID, &w.Name, &w.Code, &w.IsDefault, &w.CreatedAt)
	return &w, err
}

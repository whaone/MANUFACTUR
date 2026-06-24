//go:build integration

package material

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"manufactpro/backend/internal/auth"
	"manufactpro/backend/internal/testutil"
)

func makeToken(t *testing.T, workspaceID uuid.UUID) string {
	t.Helper()
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "dev-secret-change-in-production"
	}
	claims := auth.Claims{
		UserID:      uuid.New(),
		WorkspaceID: workspaceID,
		Role:        "owner",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}
	tok, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("makeToken: %v", err)
	}
	return tok
}

func TestMaterialCRUD(t *testing.T) {
	cleanup := testutil.Setup(t)
	defer cleanup()

	// Insert test workspace so FK is satisfied
	wsID := uuid.New()
	_, err := testutil.Pool().Exec(context.Background(),
		`INSERT INTO workspaces(id,name) VALUES($1,$2)`,
		wsID, "Test Workspace "+wsID.String())
	if err != nil {
		t.Fatalf("insert workspace: %v", err)
	}
	t.Cleanup(func() {
		testutil.Pool().Exec(context.Background(), `DELETE FROM workspaces WHERE id=$1`, wsID)
	})

	router := Routes()
	token := makeToken(t, wsID)
	auth := func(req *http.Request) *http.Request {
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		return req
	}

	// ── Create ────────────────────────────────────────────────────────────────
	sku := fmt.Sprintf("TEST-%d", time.Now().UnixMilli())
	body, _ := json.Marshal(map[string]any{
		"sku": sku, "name": "Test Material", "unit": "kg", "min_stock": 5, "is_active": true,
	})
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, auth(httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))))
	if rec.Code != http.StatusCreated {
		t.Fatalf("create: want 201, got %d — %s", rec.Code, rec.Body)
	}
	var created map[string]any
	json.Unmarshal(rec.Body.Bytes(), &created)
	id := created["id"].(string)

	// ── List ──────────────────────────────────────────────────────────────────
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, auth(httptest.NewRequest(http.MethodGet, "/", nil)))
	if rec.Code != http.StatusOK {
		t.Fatalf("list: want 200, got %d", rec.Code)
	}
	var list []map[string]any
	json.Unmarshal(rec.Body.Bytes(), &list)
	found := false
	for _, m := range list {
		if m["id"] == id {
			found = true
		}
	}
	if !found {
		t.Error("list: created material not found")
	}

	// ── Update ────────────────────────────────────────────────────────────────
	body, _ = json.Marshal(map[string]any{"name": "Updated Material"})
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, auth(httptest.NewRequest(http.MethodPut, "/"+id, bytes.NewReader(body))))
	if rec.Code != http.StatusOK {
		t.Fatalf("update: want 200, got %d — %s", rec.Code, rec.Body)
	}

	// ── Delete ────────────────────────────────────────────────────────────────
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, auth(httptest.NewRequest(http.MethodDelete, "/"+id, nil)))
	if rec.Code != http.StatusNoContent {
		t.Fatalf("delete: want 204, got %d", rec.Code)
	}
}

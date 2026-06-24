//go:build integration

package auth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"manufactpro/backend/internal/testutil"
)

func TestAuthFlow(t *testing.T) {
	cleanup := testutil.Setup(t)
	defer cleanup()

	router := Routes()
	email := fmt.Sprintf("test_%d@integration.test", time.Now().UnixMilli())

	// ── Register ──────────────────────────────────────────────────────────────
	body, _ := json.Marshal(map[string]string{
		"name":           "Integration Tester",
		"email":          email,
		"password":       "rahasia123",
		"workspace_name": "Test Workspace",
	})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("register: want 201, got %d — body: %s", rec.Code, rec.Body)
	}
	var regResp map[string]any
	json.Unmarshal(rec.Body.Bytes(), &regResp)
	if regResp["access_token"] == nil {
		t.Fatal("register: missing access_token in response")
	}
	token := regResp["access_token"].(string)

	// ── Login ─────────────────────────────────────────────────────────────────
	body, _ = json.Marshal(map[string]string{"email": email, "password": "rahasia123"})
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("login: want 200, got %d — body: %s", rec.Code, rec.Body)
	}
	var loginResp map[string]any
	json.Unmarshal(rec.Body.Bytes(), &loginResp)
	if loginResp["access_token"] == nil {
		t.Fatal("login: missing access_token")
	}
	token = loginResp["access_token"].(string)

	// ── Login wrong password ───────────────────────────────────────────────────
	body, _ = json.Marshal(map[string]string{"email": email, "password": "salah"})
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("login wrong pwd: want 401, got %d", rec.Code)
	}

	// ── Me ────────────────────────────────────────────────────────────────────
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("me: want 200, got %d — body: %s", rec.Code, rec.Body)
	}
	var meResp map[string]any
	json.Unmarshal(rec.Body.Bytes(), &meResp)
	if meResp["email"] != email {
		t.Errorf("me: want email %s, got %v", email, meResp["email"])
	}

	// ── Me without token ──────────────────────────────────────────────────────
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/me", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("me no token: want 401, got %d", rec.Code)
	}
}

package auth

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"

	"manufactpro/backend/internal/db"
)

type Claims struct {
	UserID      uuid.UUID `json:"user_id"`
	WorkspaceID uuid.UUID `json:"workspace_id"`
	Role        string    `json:"role"`
	jwt.RegisteredClaims
}

type User struct {
	ID          uuid.UUID `json:"id"`
	WorkspaceID uuid.UUID `json:"workspace_id"`
	Name        string    `json:"name"`
	Email       string    `json:"email"`
	Role        string    `json:"role"`
}

type LoginResult struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	User         User   `json:"user"`
}

func jwtSecret() []byte {
	s := os.Getenv("JWT_SECRET")
	if s == "" {
		s = "dev-secret-change-in-production"
	}
	return []byte(s)
}

func makeToken(user User, expiry time.Duration) (string, error) {
	claims := Claims{
		UserID:      user.ID,
		WorkspaceID: user.WorkspaceID,
		Role:        user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(jwtSecret())
}

func Register(ctx context.Context, name, email, password, workspaceName string) (*LoginResult, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("bcrypt: %w", err)
	}

	tx, err := db.Pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	var wsID uuid.UUID
	err = tx.QueryRow(ctx,
		`INSERT INTO workspaces(name) VALUES($1) RETURNING id`,
		workspaceName,
	).Scan(&wsID)
	if err != nil {
		return nil, fmt.Errorf("create workspace: %w", err)
	}

	user := User{WorkspaceID: wsID, Name: name, Email: email, Role: "owner"}
	err = tx.QueryRow(ctx,
		`INSERT INTO users(workspace_id,name,email,password_hash,role) VALUES($1,$2,$3,$4,'owner') RETURNING id`,
		wsID, name, email, string(hash),
	).Scan(&user.ID)
	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return buildResult(user)
}

func Login(ctx context.Context, email, password string) (*LoginResult, error) {
	var user User
	var hash string
	err := db.Pool.QueryRow(ctx,
		`SELECT id, workspace_id, name, email, role, password_hash FROM users WHERE email=$1`,
		email,
	).Scan(&user.ID, &user.WorkspaceID, &user.Name, &user.Email, &user.Role, &hash)
	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("invalid credentials")
	}
	if err != nil {
		return nil, err
	}
	if bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) != nil {
		return nil, fmt.Errorf("invalid credentials")
	}
	return buildResult(user)
}

// Refresh validates a refresh token and issues a fresh token pair.
// The user is re-loaded from DB so role/workspace changes take effect and
// deleted users are rejected.
func Refresh(ctx context.Context, refreshToken string) (*LoginResult, error) {
	claims, err := ParseToken(refreshToken)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token")
	}
	u, err := Me(ctx, claims.UserID)
	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("user no longer exists")
	}
	if err != nil {
		return nil, err
	}
	return buildResult(*u)
}

func Me(ctx context.Context, userID uuid.UUID) (*User, error) {
	var u User
	err := db.Pool.QueryRow(ctx,
		`SELECT id, workspace_id, name, email, role FROM users WHERE id=$1`,
		userID,
	).Scan(&u.ID, &u.WorkspaceID, &u.Name, &u.Email, &u.Role)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func buildResult(user User) (*LoginResult, error) {
	access, err := makeToken(user, 24*time.Hour)
	if err != nil {
		return nil, err
	}
	refresh, err := makeToken(user, 7*24*time.Hour)
	if err != nil {
		return nil, err
	}
	return &LoginResult{AccessToken: access, RefreshToken: refresh, User: user}, nil
}

func ParseToken(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return jwtSecret(), nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	return claims, nil
}

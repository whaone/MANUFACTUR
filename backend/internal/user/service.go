package user

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"manufactpro/backend/internal/db"
)

type User struct {
	ID          uuid.UUID `json:"id"`
	WorkspaceID uuid.UUID `json:"workspace_id"`
	Name        string    `json:"name"`
	Email       string    `json:"email"`
	Role        string    `json:"role"`
	CreatedAt   time.Time `json:"created_at"`
}

var validRoles = map[string]bool{
	"owner": true, "admin": true, "production": true,
	"warehouse": true, "viewer": true,
}

func List(ctx context.Context, workspaceID uuid.UUID) ([]User, error) {
	rows, err := db.Pool.Query(ctx,
		`SELECT id, workspace_id, name, email, role, created_at
		 FROM users WHERE workspace_id=$1 ORDER BY created_at`,
		workspaceID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.WorkspaceID, &u.Name, &u.Email, &u.Role, &u.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, u)
	}
	if list == nil {
		list = []User{}
	}
	return list, nil
}

func UpdateRole(ctx context.Context, workspaceID, userID uuid.UUID, role string) (*User, error) {
	if !validRoles[role] {
		return nil, fmt.Errorf("invalid role: %s", role)
	}
	var u User
	err := db.Pool.QueryRow(ctx,
		`UPDATE users SET role=$1 WHERE id=$2 AND workspace_id=$3
		 RETURNING id, workspace_id, name, email, role, created_at`,
		role, userID, workspaceID,
	).Scan(&u.ID, &u.WorkspaceID, &u.Name, &u.Email, &u.Role, &u.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}
	return &u, nil
}

func InviteUser(ctx context.Context, workspaceID uuid.UUID, name, email, role, passwordHash string) (*User, error) {
	if !validRoles[role] {
		return nil, fmt.Errorf("invalid role")
	}
	var u User
	err := db.Pool.QueryRow(ctx,
		`INSERT INTO users(workspace_id,name,email,password_hash,role)
		 VALUES($1,$2,$3,$4,$5)
		 RETURNING id, workspace_id, name, email, role, created_at`,
		workspaceID, name, email, passwordHash, role,
	).Scan(&u.ID, &u.WorkspaceID, &u.Name, &u.Email, &u.Role, &u.CreatedAt)
	return &u, err
}

func DeleteUser(ctx context.Context, workspaceID, userID uuid.UUID) error {
	_, err := db.Pool.Exec(ctx,
		`DELETE FROM users WHERE id=$1 AND workspace_id=$2`,
		userID, workspaceID,
	)
	return err
}

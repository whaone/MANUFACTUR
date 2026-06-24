package supplier

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"manufactpro/backend/internal/db"
)

type Supplier struct {
	ID          uuid.UUID `json:"id"`
	WorkspaceID uuid.UUID `json:"workspace_id"`
	Name        string    `json:"name"`
	Contact     string    `json:"contact"`
	Email       string    `json:"email"`
	Phone       string    `json:"phone"`
	PaymentTerm string    `json:"payment_term"`
	CreatedAt   time.Time `json:"created_at"`
}

type CreateInput struct {
	Name        string `json:"name"`
	Contact     string `json:"contact"`
	Email       string `json:"email"`
	Phone       string `json:"phone"`
	PaymentTerm string `json:"payment_term"`
}

type UpdateInput struct {
	Name        *string `json:"name"`
	Contact     *string `json:"contact"`
	Email       *string `json:"email"`
	Phone       *string `json:"phone"`
	PaymentTerm *string `json:"payment_term"`
}

func scan(row interface{ Scan(...any) error }) (*Supplier, error) {
	var s Supplier
	err := row.Scan(&s.ID, &s.WorkspaceID, &s.Name,
		&s.Contact, &s.Email, &s.Phone, &s.PaymentTerm, &s.CreatedAt)
	return &s, err
}

const selectCols = `id, workspace_id, name,
	COALESCE(contact,''), COALESCE(email,''), COALESCE(phone,''), COALESCE(payment_term,''), created_at`

func List(ctx context.Context, workspaceID uuid.UUID) ([]Supplier, error) {
	rows, err := db.Pool.Query(ctx,
		`SELECT `+selectCols+` FROM suppliers WHERE workspace_id=$1 ORDER BY name`,
		workspaceID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []Supplier
	for rows.Next() {
		s, err := scan(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, *s)
	}
	if list == nil {
		list = []Supplier{}
	}
	return list, nil
}

func GetByID(ctx context.Context, workspaceID, id uuid.UUID) (*Supplier, error) {
	return scan(db.Pool.QueryRow(ctx,
		`SELECT `+selectCols+` FROM suppliers WHERE id=$1 AND workspace_id=$2`,
		id, workspaceID,
	))
}

func Create(ctx context.Context, workspaceID uuid.UUID, in CreateInput) (*Supplier, error) {
	if in.Name == "" {
		return nil, fmt.Errorf("name required")
	}
	return scan(db.Pool.QueryRow(ctx,
		`INSERT INTO suppliers(workspace_id,name,contact,email,phone,payment_term)
		 VALUES($1,$2,$3,$4,$5,$6)
		 RETURNING `+selectCols,
		workspaceID, in.Name, in.Contact, in.Email, in.Phone, in.PaymentTerm,
	))
}

func Update(ctx context.Context, workspaceID, id uuid.UUID, in UpdateInput) (*Supplier, error) {
	cur, err := GetByID(ctx, workspaceID, id)
	if err != nil {
		return nil, err
	}
	if in.Name != nil { cur.Name = *in.Name }
	if in.Contact != nil { cur.Contact = *in.Contact }
	if in.Email != nil { cur.Email = *in.Email }
	if in.Phone != nil { cur.Phone = *in.Phone }
	if in.PaymentTerm != nil { cur.PaymentTerm = *in.PaymentTerm }

	return scan(db.Pool.QueryRow(ctx,
		`UPDATE suppliers SET name=$1,contact=$2,email=$3,phone=$4,payment_term=$5
		 WHERE id=$6 AND workspace_id=$7 RETURNING `+selectCols,
		cur.Name, cur.Contact, cur.Email, cur.Phone, cur.PaymentTerm, id, workspaceID,
	))
}

func Delete(ctx context.Context, workspaceID, id uuid.UUID) error {
	_, err := db.Pool.Exec(ctx,
		`DELETE FROM suppliers WHERE id=$1 AND workspace_id=$2`, id, workspaceID,
	)
	return err
}

package material

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"manufactpro/backend/internal/db"
)

type Material struct {
	ID          uuid.UUID  `json:"id"`
	WorkspaceID uuid.UUID  `json:"workspace_id"`
	SKU         string     `json:"sku"`
	Name        string     `json:"name"`
	Unit        string     `json:"unit"`
	Category    string     `json:"category"`
	MinStock    float64    `json:"min_stock"`
	Barcode     string     `json:"barcode"`
	ImageURL    string     `json:"image_url"`
	IsActive    bool       `json:"is_active"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type CreateInput struct {
	SKU      string  `json:"sku"`
	Name     string  `json:"name"`
	Unit     string  `json:"unit"`
	Category string  `json:"category"`
	MinStock float64 `json:"min_stock"`
	Barcode  string  `json:"barcode"`
	ImageURL string  `json:"image_url"`
	IsActive bool    `json:"is_active"`
}

type UpdateInput struct {
	SKU      *string  `json:"sku"`
	Name     *string  `json:"name"`
	Unit     *string  `json:"unit"`
	Category *string  `json:"category"`
	MinStock *float64 `json:"min_stock"`
	Barcode  *string  `json:"barcode"`
	ImageURL *string  `json:"image_url"`
	IsActive *bool    `json:"is_active"`
}

func List(ctx context.Context, workspaceID uuid.UUID) ([]Material, error) {
	rows, err := db.Pool.Query(ctx,
		`SELECT id, workspace_id, sku, name, unit, category, min_stock,
		        COALESCE(barcode,''), COALESCE(image_url,''), is_active, created_at, updated_at
		 FROM materials WHERE workspace_id=$1 AND is_active=true ORDER BY name`,
		workspaceID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []Material
	for rows.Next() {
		var m Material
		if err := rows.Scan(&m.ID, &m.WorkspaceID, &m.SKU, &m.Name, &m.Unit, &m.Category,
			&m.MinStock, &m.Barcode, &m.ImageURL, &m.IsActive, &m.CreatedAt, &m.UpdatedAt); err != nil {
			return nil, err
		}
		list = append(list, m)
	}
	if list == nil {
		list = []Material{}
	}
	return list, nil
}

func GetByID(ctx context.Context, workspaceID, id uuid.UUID) (*Material, error) {
	var m Material
	err := db.Pool.QueryRow(ctx,
		`SELECT id, workspace_id, sku, name, unit, category, min_stock,
		        COALESCE(barcode,''), COALESCE(image_url,''), is_active, created_at, updated_at
		 FROM materials WHERE id=$1 AND workspace_id=$2`,
		id, workspaceID,
	).Scan(&m.ID, &m.WorkspaceID, &m.SKU, &m.Name, &m.Unit, &m.Category,
		&m.MinStock, &m.Barcode, &m.ImageURL, &m.IsActive, &m.CreatedAt, &m.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func Create(ctx context.Context, workspaceID uuid.UUID, in CreateInput) (*Material, error) {
	if in.SKU == "" || in.Name == "" || in.Unit == "" {
		return nil, fmt.Errorf("sku, name, unit required")
	}
	var m Material
	err := db.Pool.QueryRow(ctx,
		`INSERT INTO materials(workspace_id,sku,name,unit,category,min_stock,barcode,image_url,is_active)
		 VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9)
		 RETURNING id, workspace_id, sku, name, unit, category, min_stock,
		           COALESCE(barcode,''), COALESCE(image_url,''), is_active, created_at, updated_at`,
		workspaceID, in.SKU, in.Name, in.Unit, in.Category, in.MinStock, in.Barcode, in.ImageURL, in.IsActive,
	).Scan(&m.ID, &m.WorkspaceID, &m.SKU, &m.Name, &m.Unit, &m.Category,
		&m.MinStock, &m.Barcode, &m.ImageURL, &m.IsActive, &m.CreatedAt, &m.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func Update(ctx context.Context, workspaceID, id uuid.UUID, in UpdateInput) (*Material, error) {
	cur, err := GetByID(ctx, workspaceID, id)
	if err != nil {
		return nil, err
	}
	if in.SKU != nil { cur.SKU = *in.SKU }
	if in.Name != nil { cur.Name = *in.Name }
	if in.Unit != nil { cur.Unit = *in.Unit }
	if in.Category != nil { cur.Category = *in.Category }
	if in.MinStock != nil { cur.MinStock = *in.MinStock }
	if in.Barcode != nil { cur.Barcode = *in.Barcode }
	if in.ImageURL != nil { cur.ImageURL = *in.ImageURL }
	if in.IsActive != nil { cur.IsActive = *in.IsActive }

	var m Material
	err = db.Pool.QueryRow(ctx,
		`UPDATE materials SET sku=$1,name=$2,unit=$3,category=$4,min_stock=$5,
		        barcode=$6,image_url=$7,is_active=$8,updated_at=NOW()
		 WHERE id=$9 AND workspace_id=$10
		 RETURNING id, workspace_id, sku, name, unit, category, min_stock,
		           COALESCE(barcode,''), COALESCE(image_url,''), is_active, created_at, updated_at`,
		cur.SKU, cur.Name, cur.Unit, cur.Category, cur.MinStock,
		cur.Barcode, cur.ImageURL, cur.IsActive, id, workspaceID,
	).Scan(&m.ID, &m.WorkspaceID, &m.SKU, &m.Name, &m.Unit, &m.Category,
		&m.MinStock, &m.Barcode, &m.ImageURL, &m.IsActive, &m.CreatedAt, &m.UpdatedAt)
	return &m, err
}

func Delete(ctx context.Context, workspaceID, id uuid.UUID) error {
	_, err := db.Pool.Exec(ctx,
		`UPDATE materials SET is_active=false, updated_at=NOW() WHERE id=$1 AND workspace_id=$2`,
		id, workspaceID,
	)
	return err
}

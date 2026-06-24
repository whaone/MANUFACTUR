package product

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"manufactpro/backend/internal/db"
)

type Variant struct {
	ID         uuid.UUID         `json:"id"`
	ProductID  uuid.UUID         `json:"product_id"`
	SKU        string            `json:"sku"`
	Barcode    string            `json:"barcode"`
	Attributes map[string]string `json:"attributes"`
	SellPrice  float64           `json:"sell_price"`
	IsActive   bool              `json:"is_active"`
	CreatedAt  time.Time         `json:"created_at"`
}

type Product struct {
	ID          uuid.UUID `json:"id"`
	WorkspaceID uuid.UUID `json:"workspace_id"`
	Name        string    `json:"name"`
	Category    string    `json:"category"`
	Description string    `json:"description"`
	ImageURL    string    `json:"image_url"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Variants    []Variant `json:"variants"`
}

type CreateProductInput struct {
	Name        string `json:"name"`
	Category    string `json:"category"`
	Description string `json:"description"`
	ImageURL    string `json:"image_url"`
}

type UpdateProductInput struct {
	Name        *string `json:"name"`
	Category    *string `json:"category"`
	Description *string `json:"description"`
	ImageURL    *string `json:"image_url"`
}

type CreateVariantInput struct {
	SKU        string            `json:"sku"`
	Barcode    string            `json:"barcode"`
	Attributes map[string]string `json:"attributes"`
	SellPrice  float64           `json:"sell_price"`
	IsActive   bool              `json:"is_active"`
}

type UpdateVariantInput struct {
	SKU        *string           `json:"sku"`
	Barcode    *string           `json:"barcode"`
	Attributes map[string]string `json:"attributes"`
	SellPrice  *float64          `json:"sell_price"`
	IsActive   *bool             `json:"is_active"`
}

func scanVariant(row pgx.Row) (*Variant, error) {
	var v Variant
	var attrsJSON []byte
	err := row.Scan(&v.ID, &v.ProductID, &v.SKU, &v.Barcode, &attrsJSON, &v.SellPrice, &v.IsActive, &v.CreatedAt)
	if err != nil {
		return nil, err
	}
	json.Unmarshal(attrsJSON, &v.Attributes)
	if v.Attributes == nil {
		v.Attributes = map[string]string{}
	}
	return &v, nil
}

func listVariants(ctx context.Context, productID uuid.UUID) ([]Variant, error) {
	rows, err := db.Pool.Query(ctx,
		`SELECT id, product_id, sku, COALESCE(barcode,''), attributes, sell_price, is_active, created_at
		 FROM product_variants WHERE product_id=$1 ORDER BY sku`,
		productID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []Variant
	for rows.Next() {
		var v Variant
		var attrsJSON []byte
		if err := rows.Scan(&v.ID, &v.ProductID, &v.SKU, &v.Barcode, &attrsJSON, &v.SellPrice, &v.IsActive, &v.CreatedAt); err != nil {
			return nil, err
		}
		json.Unmarshal(attrsJSON, &v.Attributes)
		if v.Attributes == nil {
			v.Attributes = map[string]string{}
		}
		list = append(list, v)
	}
	if list == nil {
		list = []Variant{}
	}
	return list, nil
}

func List(ctx context.Context, workspaceID uuid.UUID) ([]Product, error) {
	rows, err := db.Pool.Query(ctx,
		`SELECT id, workspace_id, name, COALESCE(category,''), COALESCE(description,''), COALESCE(image_url,''), created_at, updated_at
		 FROM products WHERE workspace_id=$1 ORDER BY name`,
		workspaceID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []Product
	for rows.Next() {
		var p Product
		if err := rows.Scan(&p.ID, &p.WorkspaceID, &p.Name, &p.Category, &p.Description, &p.ImageURL, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		list = append(list, p)
	}
	if list == nil {
		return []Product{}, nil
	}

	// Attach variants without an N+1: fetch them all in one query, group by product.
	variants, err := ListVariantsFlat(ctx, workspaceID)
	if err != nil {
		return nil, err
	}
	byProduct := make(map[uuid.UUID][]Variant, len(list))
	for _, v := range variants {
		byProduct[v.ProductID] = append(byProduct[v.ProductID], v)
	}
	for i := range list {
		if vs := byProduct[list[i].ID]; vs != nil {
			list[i].Variants = vs
		} else {
			list[i].Variants = []Variant{}
		}
	}
	return list, nil
}

func GetByID(ctx context.Context, workspaceID, id uuid.UUID) (*Product, error) {
	var p Product
	err := db.Pool.QueryRow(ctx,
		`SELECT id, workspace_id, name, COALESCE(category,''), COALESCE(description,''), COALESCE(image_url,''), created_at, updated_at
		 FROM products WHERE id=$1 AND workspace_id=$2`,
		id, workspaceID,
	).Scan(&p.ID, &p.WorkspaceID, &p.Name, &p.Category, &p.Description, &p.ImageURL, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, err
	}
	p.Variants, _ = listVariants(ctx, p.ID)
	return &p, nil
}

func Create(ctx context.Context, workspaceID uuid.UUID, in CreateProductInput) (*Product, error) {
	if in.Name == "" {
		return nil, fmt.Errorf("name required")
	}
	var p Product
	err := db.Pool.QueryRow(ctx,
		`INSERT INTO products(workspace_id,name,category,description,image_url) VALUES($1,$2,$3,$4,$5)
		 RETURNING id, workspace_id, name, COALESCE(category,''), COALESCE(description,''), COALESCE(image_url,''), created_at, updated_at`,
		workspaceID, in.Name, in.Category, in.Description, in.ImageURL,
	).Scan(&p.ID, &p.WorkspaceID, &p.Name, &p.Category, &p.Description, &p.ImageURL, &p.CreatedAt, &p.UpdatedAt)
	p.Variants = []Variant{}
	return &p, err
}

func Update(ctx context.Context, workspaceID, id uuid.UUID, in UpdateProductInput) (*Product, error) {
	cur, err := GetByID(ctx, workspaceID, id)
	if err != nil {
		return nil, err
	}
	if in.Name != nil {
		cur.Name = *in.Name
	}
	if in.Category != nil {
		cur.Category = *in.Category
	}
	if in.Description != nil {
		cur.Description = *in.Description
	}
	if in.ImageURL != nil {
		cur.ImageURL = *in.ImageURL
	}

	var p Product
	err = db.Pool.QueryRow(ctx,
		`UPDATE products SET name=$1,category=$2,description=$3,image_url=$4,updated_at=NOW()
		 WHERE id=$5 AND workspace_id=$6
		 RETURNING id, workspace_id, name, COALESCE(category,''), COALESCE(description,''), COALESCE(image_url,''), created_at, updated_at`,
		cur.Name, cur.Category, cur.Description, cur.ImageURL, id, workspaceID,
	).Scan(&p.ID, &p.WorkspaceID, &p.Name, &p.Category, &p.Description, &p.ImageURL, &p.CreatedAt, &p.UpdatedAt)
	p.Variants, _ = listVariants(ctx, p.ID)
	return &p, err
}

func Delete(ctx context.Context, workspaceID, id uuid.UUID) error {
	tx, err := db.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Verify ownership first.
	var exists bool
	if err := tx.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM products WHERE id=$1 AND workspace_id=$2)`,
		id, workspaceID,
	).Scan(&exists); err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("product not found")
	}

	// Cascade: bom_items → product_variants → product (avoid FK violation).
	if _, err := tx.Exec(ctx,
		`DELETE FROM bom_items b USING product_variants pv
		 WHERE b.product_variant_id=pv.id AND pv.product_id=$1`, id,
	); err != nil {
		return fmt.Errorf("delete bom_items: %w", err)
	}
	if _, err := tx.Exec(ctx,
		`DELETE FROM product_variants WHERE product_id=$1`, id,
	); err != nil {
		return fmt.Errorf("delete variants: %w", err)
	}
	if _, err := tx.Exec(ctx,
		`DELETE FROM products WHERE id=$1 AND workspace_id=$2`, id, workspaceID,
	); err != nil {
		return fmt.Errorf("delete product: %w", err)
	}
	return tx.Commit(ctx)
}

func ListVariantsFlat(ctx context.Context, workspaceID uuid.UUID) ([]Variant, error) {
	rows, err := db.Pool.Query(ctx,
		`SELECT pv.id, pv.product_id, pv.sku, COALESCE(pv.barcode,''), pv.attributes, pv.sell_price, pv.is_active, pv.created_at
		 FROM product_variants pv
		 JOIN products p ON p.id = pv.product_id
		 WHERE p.workspace_id=$1 ORDER BY pv.sku`,
		workspaceID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []Variant
	for rows.Next() {
		var v Variant
		var attrsJSON []byte
		if err := rows.Scan(&v.ID, &v.ProductID, &v.SKU, &v.Barcode, &attrsJSON, &v.SellPrice, &v.IsActive, &v.CreatedAt); err != nil {
			return nil, err
		}
		json.Unmarshal(attrsJSON, &v.Attributes)
		if v.Attributes == nil {
			v.Attributes = map[string]string{}
		}
		list = append(list, v)
	}
	if list == nil {
		list = []Variant{}
	}
	return list, nil
}

func CreateVariant(ctx context.Context, workspaceID, productID uuid.UUID, in CreateVariantInput) (*Variant, error) {
	// verify product belongs to workspace
	var count int
	db.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM products WHERE id=$1 AND workspace_id=$2`, productID, workspaceID).Scan(&count)
	if count == 0 {
		return nil, fmt.Errorf("product not found")
	}
	if in.SKU == "" {
		return nil, fmt.Errorf("sku required")
	}
	attrsJSON, _ := json.Marshal(in.Attributes)
	return scanVariant(db.Pool.QueryRow(ctx,
		`INSERT INTO product_variants(product_id,sku,barcode,attributes,sell_price,is_active) VALUES($1,$2,$3,$4,$5,$6)
		 RETURNING id, product_id, sku, COALESCE(barcode,''), attributes, sell_price, is_active, created_at`,
		productID, in.SKU, in.Barcode, attrsJSON, in.SellPrice, in.IsActive,
	))
}

func UpdateVariant(ctx context.Context, workspaceID, productID, variantID uuid.UUID, in UpdateVariantInput) (*Variant, error) {
	var v Variant
	var attrsJSON []byte
	err := db.Pool.QueryRow(ctx,
		`SELECT pv.id, pv.product_id, pv.sku, COALESCE(pv.barcode,''), pv.attributes, pv.sell_price, pv.is_active, pv.created_at
		 FROM product_variants pv JOIN products p ON p.id=pv.product_id
		 WHERE pv.id=$1 AND pv.product_id=$2 AND p.workspace_id=$3`,
		variantID, productID, workspaceID,
	).Scan(&v.ID, &v.ProductID, &v.SKU, &v.Barcode, &attrsJSON, &v.SellPrice, &v.IsActive, &v.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("variant not found")
	}
	json.Unmarshal(attrsJSON, &v.Attributes)

	if in.SKU != nil {
		v.SKU = *in.SKU
	}
	if in.Barcode != nil {
		v.Barcode = *in.Barcode
	}
	if in.SellPrice != nil {
		v.SellPrice = *in.SellPrice
	}
	if in.IsActive != nil {
		v.IsActive = *in.IsActive
	}
	if in.Attributes != nil {
		v.Attributes = in.Attributes
	}

	newAttrs, _ := json.Marshal(v.Attributes)
	return scanVariant(db.Pool.QueryRow(ctx,
		`UPDATE product_variants SET sku=$1,barcode=$2,attributes=$3,sell_price=$4,is_active=$5
		 WHERE id=$6 RETURNING id, product_id, sku, COALESCE(barcode,''), attributes, sell_price, is_active, created_at`,
		v.SKU, v.Barcode, newAttrs, v.SellPrice, v.IsActive, variantID,
	))
}

func DeleteVariant(ctx context.Context, workspaceID, productID, variantID uuid.UUID) error {
	_, err := db.Pool.Exec(ctx,
		`DELETE FROM product_variants pv USING products p
		 WHERE p.id=pv.product_id AND pv.id=$1 AND pv.product_id=$2 AND p.workspace_id=$3`,
		variantID, productID, workspaceID,
	)
	return err
}

func UpdateVariantByID(ctx context.Context, workspaceID, variantID uuid.UUID, in UpdateVariantInput) (*Variant, error) {
	var v Variant
	var attrsJSON []byte
	err := db.Pool.QueryRow(ctx,
		`SELECT pv.id, pv.product_id, pv.sku, COALESCE(pv.barcode,''), pv.attributes, pv.sell_price, pv.is_active, pv.created_at
		 FROM product_variants pv JOIN products p ON p.id=pv.product_id
		 WHERE pv.id=$1 AND p.workspace_id=$2`,
		variantID, workspaceID,
	).Scan(&v.ID, &v.ProductID, &v.SKU, &v.Barcode, &attrsJSON, &v.SellPrice, &v.IsActive, &v.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("variant not found")
	}
	json.Unmarshal(attrsJSON, &v.Attributes)

	if in.SKU != nil {
		v.SKU = *in.SKU
	}
	if in.Barcode != nil {
		v.Barcode = *in.Barcode
	}
	if in.SellPrice != nil {
		v.SellPrice = *in.SellPrice
	}
	if in.IsActive != nil {
		v.IsActive = *in.IsActive
	}
	if in.Attributes != nil {
		v.Attributes = in.Attributes
	}

	newAttrs, _ := json.Marshal(v.Attributes)
	return scanVariant(db.Pool.QueryRow(ctx,
		`UPDATE product_variants SET sku=$1,barcode=$2,attributes=$3,sell_price=$4,is_active=$5
		 WHERE id=$6 RETURNING id, product_id, sku, COALESCE(barcode,''), attributes, sell_price, is_active, created_at`,
		v.SKU, v.Barcode, newAttrs, v.SellPrice, v.IsActive, variantID,
	))
}

func DeleteVariantByID(ctx context.Context, workspaceID, variantID uuid.UUID) error {
	_, err := db.Pool.Exec(ctx,
		`DELETE FROM product_variants pv USING products p
		 WHERE p.id=pv.product_id AND pv.id=$1 AND p.workspace_id=$2`,
		variantID, workspaceID,
	)
	return err
}

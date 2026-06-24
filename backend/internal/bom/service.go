package bom

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"manufactpro/backend/internal/db"
)

type BomItem struct {
	ID               uuid.UUID `json:"id"`
	ProductVariantID uuid.UUID `json:"product_variant_id"`
	MaterialID       uuid.UUID `json:"material_id"`
	MaterialName     string    `json:"material_name,omitempty"`
	MaterialSKU      string    `json:"material_sku,omitempty"`
	Qty              float64   `json:"qty"`
	Unit             string    `json:"unit"`
	IsOptional       bool      `json:"is_optional"`
	Note             string    `json:"note"`
}

const selectCols = `
	b.id, b.product_variant_id, b.material_id,
	m.name AS material_name, m.sku AS material_sku,
	b.qty, b.unit, b.is_optional, COALESCE(b.note, '')`

// ListByVariant returns BOM items for a variant, workspace-safe via JOIN.
func ListByVariant(ctx context.Context, workspaceID, variantID uuid.UUID) ([]BomItem, error) {
	rows, err := db.Pool.Query(ctx, `
		SELECT `+selectCols+`
		FROM bom_items b
		JOIN materials m ON m.id = b.material_id
		JOIN product_variants pv ON pv.id = b.product_variant_id
		JOIN products p ON p.id = pv.product_id
		WHERE b.product_variant_id = $1 AND p.workspace_id = $2
		ORDER BY m.name`,
		variantID, workspaceID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []BomItem
	for rows.Next() {
		var it BomItem
		if err := rows.Scan(&it.ID, &it.ProductVariantID, &it.MaterialID,
			&it.MaterialName, &it.MaterialSKU,
			&it.Qty, &it.Unit, &it.IsOptional, &it.Note); err != nil {
			return nil, err
		}
		items = append(items, it)
	}
	if items == nil {
		items = []BomItem{}
	}
	return items, nil
}

type CreateBomInput struct {
	MaterialID uuid.UUID `json:"material_id"`
	Qty        float64   `json:"qty"`
	Unit       string    `json:"unit"`
	IsOptional bool      `json:"is_optional"`
	Note       string    `json:"note"`
}

func Create(ctx context.Context, workspaceID, variantID uuid.UUID, in CreateBomInput) (*BomItem, error) {
	if in.Qty <= 0 {
		return nil, fmt.Errorf("qty must be > 0")
	}
	if in.Unit == "" {
		return nil, fmt.Errorf("unit required")
	}

	// Verify variant belongs to workspace
	var exists bool
	err := db.Pool.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM product_variants pv
			JOIN products p ON p.id = pv.product_id
			WHERE pv.id = $1 AND p.workspace_id = $2
		)`, variantID, workspaceID).Scan(&exists)
	if err != nil || !exists {
		return nil, fmt.Errorf("variant not found")
	}

	var it BomItem
	err = db.Pool.QueryRow(ctx, `
		INSERT INTO bom_items(product_variant_id, material_id, qty, unit, is_optional, note)
		VALUES($1, $2, $3, $4, $5, $6)
		RETURNING id, product_variant_id, material_id, qty, unit, is_optional, COALESCE(note, '')`,
		variantID, in.MaterialID, in.Qty, in.Unit, in.IsOptional, in.Note,
	).Scan(&it.ID, &it.ProductVariantID, &it.MaterialID, &it.Qty, &it.Unit, &it.IsOptional, &it.Note)
	if err != nil {
		return nil, fmt.Errorf("create bom_item: %w", err)
	}
	return &it, nil
}

type UpdateBomInput struct {
	Qty        *float64 `json:"qty"`
	Unit       *string  `json:"unit"`
	IsOptional *bool    `json:"is_optional"`
	Note       *string  `json:"note"`
}

func Update(ctx context.Context, workspaceID, itemID uuid.UUID, in UpdateBomInput) (*BomItem, error) {
	// Workspace-safe fetch
	var it BomItem
	err := db.Pool.QueryRow(ctx, `
		SELECT `+selectCols+`
		FROM bom_items b
		JOIN materials m ON m.id = b.material_id
		JOIN product_variants pv ON pv.id = b.product_variant_id
		JOIN products p ON p.id = pv.product_id
		WHERE b.id = $1 AND p.workspace_id = $2`,
		itemID, workspaceID,
	).Scan(&it.ID, &it.ProductVariantID, &it.MaterialID,
		&it.MaterialName, &it.MaterialSKU,
		&it.Qty, &it.Unit, &it.IsOptional, &it.Note)
	if err != nil {
		return nil, fmt.Errorf("bom_item not found")
	}

	if in.Qty != nil {
		it.Qty = *in.Qty
	}
	if in.Unit != nil {
		it.Unit = *in.Unit
	}
	if in.IsOptional != nil {
		it.IsOptional = *in.IsOptional
	}
	if in.Note != nil {
		it.Note = *in.Note
	}

	_, err = db.Pool.Exec(ctx, `
		UPDATE bom_items SET qty=$1, unit=$2, is_optional=$3, note=$4 WHERE id=$5`,
		it.Qty, it.Unit, it.IsOptional, it.Note, it.ID,
	)
	if err != nil {
		return nil, fmt.Errorf("update bom_item: %w", err)
	}
	return &it, nil
}

func Delete(ctx context.Context, workspaceID, itemID uuid.UUID) error {
	tag, err := db.Pool.Exec(ctx, `
		DELETE FROM bom_items b
		USING product_variants pv, products p
		WHERE b.id = $1
		  AND b.product_variant_id = pv.id
		  AND pv.product_id = p.id
		  AND p.workspace_id = $2`,
		itemID, workspaceID,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("bom_item not found")
	}
	return nil
}

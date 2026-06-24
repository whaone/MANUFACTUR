CREATE VIEW v_current_stock AS
SELECT
    sm.workspace_id,
    sm.warehouse_id,
    sm.item_type,
    sm.item_id,
    SUM(sm.qty)         AS qty_on_hand,
    MAX(sm.created_at)  AS last_movement_at
FROM stock_movements sm
GROUP BY sm.workspace_id, sm.warehouse_id, sm.item_type, sm.item_id;

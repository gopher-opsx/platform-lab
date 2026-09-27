INSERT INTO inventory (product_id, available_quantity, reserved_quantity)
VALUES
  ('prod-001', 25, 0),
  ('prod-002', 15, 0),
  ('prod-003', 5, 0)
ON CONFLICT (product_id) DO UPDATE
SET available_quantity = EXCLUDED.available_quantity,
    reserved_quantity = EXCLUDED.reserved_quantity,
    updated_at = NOW();

-- which rule sent this request to the desk, so a dealer sees only their own queue
ALTER TABLE hst.orders ADD COLUMN IF NOT EXISTS routing_id BIGINT NOT NULL DEFAULT 0;

CREATE INDEX IF NOT EXISTS orders_routing_idx ON hst.orders (routing_id) WHERE routing_id <> 0;

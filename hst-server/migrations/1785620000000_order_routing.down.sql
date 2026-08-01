DROP INDEX IF EXISTS hst.orders_routing_idx;
ALTER TABLE hst.orders DROP COLUMN IF EXISTS routing_id;

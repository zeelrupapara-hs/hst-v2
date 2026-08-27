ALTER TABLE hst.positions DROP COLUMN IF EXISTS tick_size;
ALTER TABLE hst.positions DROP COLUMN IF EXISTS tick_value;
ALTER TABLE hst.symbols ALTER COLUMN order_flags SET DEFAULT 0;
ALTER TABLE hst.symbols ALTER COLUMN expir_flags SET DEFAULT 0;
ALTER TABLE hst.symbols ALTER COLUMN fill_flags SET DEFAULT 0;

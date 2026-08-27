-- a position is valued with the tick size/value it was opened on, like its contract size
ALTER TABLE hst.positions ADD COLUMN IF NOT EXISTS tick_size NUMERIC(20,8) NOT NULL DEFAULT 0;
ALTER TABLE hst.positions ADD COLUMN IF NOT EXISTS tick_value NUMERIC(20,8) NOT NULL DEFAULT 0;
-- flags 0 used to mean "no restriction"; it now means none allowed (MT5 *_FLAGS_NONE), so old rows get the MT5 defaults
UPDATE hst.symbols SET order_flags = 127 WHERE order_flags = 0;
UPDATE hst.symbols SET expir_flags = 15 WHERE expir_flags = 0;
UPDATE hst.symbols SET fill_flags = 3 WHERE fill_flags = 0;
ALTER TABLE hst.symbols ALTER COLUMN order_flags SET DEFAULT 127;
ALTER TABLE hst.symbols ALTER COLUMN expir_flags SET DEFAULT 15;
ALTER TABLE hst.symbols ALTER COLUMN fill_flags SET DEFAULT 3;

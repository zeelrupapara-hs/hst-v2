-- a symbol saved with tick_flags 0 was "allow realtime" by accident; the bit is now honoured, so old rows get it set and new symbols start with it
UPDATE hst.symbols SET tick_flags = tick_flags | 1 WHERE tick_flags = 0;
ALTER TABLE hst.symbols ALTER COLUMN tick_flags SET DEFAULT 1;

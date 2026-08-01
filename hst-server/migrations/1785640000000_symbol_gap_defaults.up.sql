-- gap detection moved from a hard-coded 100 points in hst-core to per-symbol config,
-- so every symbol needs the old threshold or the routing rules that read "gapped" go dead
ALTER TABLE hst.symbols ALTER COLUMN filter_gap SET DEFAULT 100;
ALTER TABLE hst.symbols ALTER COLUMN filter_gap_ticks SET DEFAULT 3;

UPDATE hst.symbols
   SET filter_gap = 100,
       filter_gap_ticks = 3
 WHERE filter_gap = 0
   AND filter_gap_ticks = 0;

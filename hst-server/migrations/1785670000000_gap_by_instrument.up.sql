-- A gap is a jump out of proportion to the instrument, and a flat number of points is not that.
-- A hundred points is ten pips on a five digit pair and ten dollars on bitcoin, so the backfill
-- that set them all alike left the coarse, expensive instruments permanently in gap mode and
-- untradeable. Scale it to the price instead, and leave the old floor under it.
UPDATE hst.symbols s
   SET filter_gap = GREATEST(100, round(0.0015 * p.price / s.point)::int)
  FROM (
      SELECT symbol_id,
             CASE digits WHEN 1 THEN 60000.0 WHEN 2 THEN 1500.0 WHEN 3 THEN 130.0 ELSE 1.1 END AS price
        FROM hst.symbols
  ) p
 WHERE p.symbol_id = s.symbol_id AND s.point > 0;

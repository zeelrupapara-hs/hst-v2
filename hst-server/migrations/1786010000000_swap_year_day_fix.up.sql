-- swap_year_day is days-in-year for percent swap calc (360/365/366), not triple-swap weekday
UPDATE hst.symbols
   SET swap_year_day = 360
 WHERE swap_year_day >= 0
   AND swap_year_day <= 7;

ALTER TABLE hst.symbols ALTER COLUMN swap_year_day SET DEFAULT 360;

COMMENT ON COLUMN hst.symbols.swap_year_day IS
    'Days in year for percent swap calculation (360, 365, or 366)';

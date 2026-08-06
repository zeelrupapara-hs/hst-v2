ALTER TABLE hst.symbols ALTER COLUMN swap_year_day SET DEFAULT 0;

COMMENT ON COLUMN hst.symbols.swap_year_day IS
    'SwapDays triple-swap day; also days/year for percent swaps';

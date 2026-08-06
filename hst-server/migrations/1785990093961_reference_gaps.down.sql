DROP INDEX IF EXISTS hst.datafeeds_index_uidx;
ALTER TABLE hst.datafeeds DROP COLUMN IF EXISTS feed_index, DROP COLUMN IF EXISTS allow_import_symbols;
ALTER TABLE hst.groups DROP CONSTRAINT IF EXISTS groups_margin_leverage_fkey;
ALTER TABLE hst.groups
  DROP COLUMN IF EXISTS company_deposit,
  DROP COLUMN IF EXISTS company_withdrawal,
  DROP COLUMN IF EXISTS margin_leverage_id;
ALTER TABLE hst.symbols DROP COLUMN IF EXISTS tick_book_volume, DROP COLUMN IF EXISTS ie_flags;

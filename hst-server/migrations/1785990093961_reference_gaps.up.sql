-- The panel showed settings the schema had no room for: a symbol's market depth volume,
-- a group's deposit and withdrawal pages, the leverage profile a group trades under, and
-- the order data feeds are tried in. Each one is a column the reference expects.

ALTER TABLE hst.symbols
  ADD COLUMN IF NOT EXISTS tick_book_volume integer NOT NULL DEFAULT 0,
  ADD COLUMN IF NOT EXISTS ie_flags integer NOT NULL DEFAULT 0;

COMMENT ON COLUMN hst.symbols.tick_book_volume IS 'market depth volume, 0 = default';
COMMENT ON COLUMN hst.symbols.ie_flags IS 'instant execution flags, e.g. fast confirmation';

ALTER TABLE hst.groups
  ADD COLUMN IF NOT EXISTS company_deposit varchar(255) NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS company_withdrawal varchar(255) NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS margin_leverage_id bigint;

COMMENT ON COLUMN hst.groups.company_deposit IS 'deposit page, empty disables the terminal command';
COMMENT ON COLUMN hst.groups.company_withdrawal IS 'withdrawal page, empty disables the command';
COMMENT ON COLUMN hst.groups.margin_leverage_id IS 'floating leverage profile, null = none';

ALTER TABLE hst.groups
  ADD CONSTRAINT groups_margin_leverage_fkey
  FOREIGN KEY (margin_leverage_id) REFERENCES hst.leverages(leverage_id) ON DELETE SET NULL;

ALTER TABLE hst.datafeeds
  ADD COLUMN IF NOT EXISTS feed_index integer NOT NULL DEFAULT 0,
  ADD COLUMN IF NOT EXISTS allow_import_symbols smallint NOT NULL DEFAULT 0;

COMMENT ON COLUMN hst.datafeeds.feed_index IS 'the order feeds are tried in, lowest first';
COMMENT ON COLUMN hst.datafeeds.allow_import_symbols IS 'import symbol settings from this feed';

-- existing feeds keep the order they were created in
WITH ordered AS (
  SELECT datafeed_id, row_number() OVER (ORDER BY datafeed_id) - 1 AS idx FROM hst.datafeeds
)
UPDATE hst.datafeeds d SET feed_index = o.idx FROM ordered o WHERE o.datafeed_id = d.datafeed_id;

CREATE UNIQUE INDEX IF NOT EXISTS datafeeds_index_uidx ON hst.datafeeds (feed_index);

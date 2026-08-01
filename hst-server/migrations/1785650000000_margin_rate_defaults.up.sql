-- A rate multiplies the margin a formula produced, and zero means no margin at all for that type.
-- Every column defaulted to zero, so reading them literally would have freed every position on the
-- platform. Market orders default to the full charge; pending orders to none, which is the charge
-- they have always carried here and what MT5 means by leaving the rate at zero.
DO $$
DECLARE c text;
BEGIN
  FOREACH c IN ARRAY ARRAY['margin_initial_buy','margin_initial_sell'] LOOP
    EXECUTE format('ALTER TABLE hst.symbols ALTER COLUMN %I SET DEFAULT 1', c);
    EXECUTE format('UPDATE hst.symbols SET %I = 1 WHERE %I = 0', c, c);
  END LOOP;
END $$;

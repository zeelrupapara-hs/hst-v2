-- A group created by hand in SQL came out different from one created through the API: no
-- connection permission, no news or mail, no swaps, and a stop out level of zero. The API now
-- states one default table; these are the same values, so both doors lead to the same group.

ALTER TABLE hst.groups
  ALTER COLUMN permission_flags   SET DEFAULT 466,
  ALTER COLUMN auth_password_min  SET DEFAULT 8,
  ALTER COLUMN news_mode          SET DEFAULT 2,
  ALTER COLUMN mail_mode          SET DEFAULT 1,
  ALTER COLUMN trade_flags        SET DEFAULT 23,
  ALTER COLUMN margin_free_mode   SET DEFAULT 1,
  ALTER COLUMN margin_call        SET DEFAULT 50,
  ALTER COLUMN margin_stop_out    SET DEFAULT 30;

COMMENT ON COLUMN hst.groups.permission_flags IS 'EnPermissionsFlags bitmask; 0 = disabled, default = connection|risk warning|deal, order and balance notifications';
COMMENT ON COLUMN hst.groups.trade_flags IS 'default = swaps|trailing|experts|all signals';

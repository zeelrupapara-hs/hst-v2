ALTER TABLE hst.groups
  ALTER COLUMN permission_flags   SET DEFAULT 1,
  ALTER COLUMN auth_password_min  SET DEFAULT 0,
  ALTER COLUMN news_mode          SET DEFAULT 0,
  ALTER COLUMN mail_mode          SET DEFAULT 0,
  ALTER COLUMN trade_flags        SET DEFAULT 0,
  ALTER COLUMN margin_free_mode   SET DEFAULT 0,
  ALTER COLUMN margin_call        SET DEFAULT 0,
  ALTER COLUMN margin_stop_out    SET DEFAULT 0;

COMMENT ON COLUMN hst.groups.permission_flags IS 'EnPermissionsFlags bitmask; 0 = disabled';
COMMENT ON COLUMN hst.groups.trade_flags IS NULL;

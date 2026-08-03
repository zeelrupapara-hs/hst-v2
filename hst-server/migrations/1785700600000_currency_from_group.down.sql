ALTER TABLE hst.accounts ADD COLUMN IF NOT EXISTS currency_digits INT NOT NULL DEFAULT 2;

UPDATE hst.accounts a
   SET currency_digits = COALESCE(g.currency_digits, 2)
  FROM hst.users u
  LEFT JOIN hst.groups g ON g."group" = u."group"
 WHERE u.login = a.login;

-- a feed login is an identifier in the counterparty's system, not a number
ALTER TABLE hst.datafeeds
    ALTER COLUMN feed_login TYPE VARCHAR(64)
        USING CASE WHEN feed_login = 0 THEN '' ELSE feed_login::text END,
    ALTER COLUMN feed_login SET DEFAULT '',
    ALTER COLUMN gateway_login TYPE VARCHAR(64)
        USING CASE WHEN gateway_login = 0 THEN '' ELSE gateway_login::text END,
    ALTER COLUMN gateway_login SET DEFAULT '';

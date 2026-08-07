ALTER TABLE datafeeds
    ALTER COLUMN feed_login TYPE BIGINT
        USING COALESCE(NULLIF(regexp_replace(feed_login, '\D', '', 'g'), ''), '0')::bigint,
    ALTER COLUMN feed_login SET DEFAULT 0,
    ALTER COLUMN gateway_login TYPE BIGINT
        USING COALESCE(NULLIF(regexp_replace(gateway_login, '\D', '', 'g'), ''), '0')::bigint,
    ALTER COLUMN gateway_login SET DEFAULT 0;

-- swaps, trailing, experts, expiration, signals: what a group allows unless it says otherwise
ALTER TABLE hst.groups ALTER COLUMN trade_flags SET DEFAULT 63;

UPDATE hst.groups SET trade_flags = 63 WHERE trade_flags = 0;

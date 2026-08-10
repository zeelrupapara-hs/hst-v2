-- The journal type column moves from the logger taxonomy to the module taxonomy.

ALTER TABLE hst.journal DROP CONSTRAINT journal_type;

-- old 6=trade -> 5, old 5=user -> 1, old 3=net -> 12, the rest is server internal
UPDATE hst.journal
   SET type = CASE type WHEN 6 THEN 5 WHEN 5 THEN 1 WHEN 3 THEN 12 ELSE 0 END;

ALTER TABLE hst.journal ADD CONSTRAINT journal_type CHECK (type BETWEEN 0 AND 12);

COMMENT ON COLUMN hst.journal.type IS
    'module: 0=system 1=accounts 2=clients 3=groups 4=symbols 5=trade 6=managers 7=routing 8=datafeeds 9=mail 10=holidays 11=leverage 12=auth';

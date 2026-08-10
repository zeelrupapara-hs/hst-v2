-- Back to the logger taxonomy, best effort: modules without an old bucket fold into 0.

ALTER TABLE hst.journal DROP CONSTRAINT journal_type;

UPDATE hst.journal
   SET type = CASE type WHEN 5 THEN 6 WHEN 1 THEN 5 WHEN 12 THEN 3 ELSE 0 END;

ALTER TABLE hst.journal ADD CONSTRAINT journal_type CHECK (type BETWEEN 0 AND 8);

COMMENT ON COLUMN hst.journal.type IS
    'LogType: 0=all 1=cfg 2=sys 3=net 4=hst 5=user 6=trade 7=api 8=notify';

DROP INDEX IF EXISTS hst.groups_group_key;

ALTER TABLE hst.groups ADD COLUMN IF NOT EXISTS root BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE hst.groups ADD COLUMN IF NOT EXISTS parent_id BIGINT NULL
    REFERENCES hst.groups (group_id) ON DELETE RESTRICT;

-- rebuild the links from the paths they were derived from
UPDATE hst.groups c
   SET parent_id = p.group_id
  FROM hst.groups p
 WHERE p."group" = left(c."group", length(c."group") - position('\' in reverse(c."group")));

UPDATE hst.groups SET root = (parent_id IS NULL);

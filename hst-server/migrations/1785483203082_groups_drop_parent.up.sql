-- A group's identity is its path. There is no parent record and no child
-- record: demo\demo2\test is one group whose name happens to contain two
-- separators, and "demo2" is a section only because a path passes through it.
--
-- parent_id and root stored the same hierarchy a second time, and a second
-- copy of a fact is a copy that can disagree. The path is also the manager's
-- access mask and the websocket subject now, so it has to be the only answer.
ALTER TABLE hst.groups DROP CONSTRAINT IF EXISTS groups_parent_id_fkey;
ALTER TABLE hst.groups DROP COLUMN IF EXISTS parent_id;
ALTER TABLE hst.groups DROP COLUMN IF EXISTS root;

-- the path is the identity, so it has to be unique and worth an index: every
-- group access check and every subtree lookup starts from it
CREATE UNIQUE INDEX IF NOT EXISTS groups_group_key ON hst.groups ("group");

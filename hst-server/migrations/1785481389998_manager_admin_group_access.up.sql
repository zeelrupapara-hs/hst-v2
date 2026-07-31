-- Group access is now enforced: the masks in hst.managers.groups decide what a
-- manager may list and which websocket events reach it. Before this it was
-- stored and never read, so an empty column was harmless.
--
-- It is not harmless now. An empty column means no access to any group, which
-- is the safe reading and the intended one: a manager sees a tree because
-- somebody granted it, never by default.
--
-- The administrator is the exception. It is seeded with every right and with
-- its own group as its only mask, which would leave the one login that is
-- supposed to see everything able to list nothing at all.
UPDATE hst.managers
   SET groups = ARRAY['*']
 WHERE right_admin = 1
   -- only rows nobody has configured: an administrator deliberately narrowed
   -- to one tree stays narrowed, because that was a decision
   AND (groups = '{}' OR groups = ARRAY['managers\admin']);

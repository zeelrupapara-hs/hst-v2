-- Not reversible. Once the administrators hold '*' there is no way to tell
-- which rows this migration changed and which were already unrestricted, and
-- clearing them would lock out an administrator configured since.
--
-- To undo it for one login, set the mask you actually want:
--   UPDATE hst.managers SET groups = ARRAY['real\*'] WHERE login = 1000;
SELECT 1;

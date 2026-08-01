-- a group with no connection bit would refuse every login in it, which is not what a default means
ALTER TABLE hst.groups ALTER COLUMN permission_flags SET DEFAULT 3;

UPDATE hst.groups SET permission_flags = permission_flags | 2;

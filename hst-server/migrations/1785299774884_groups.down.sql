DROP INDEX IF EXISTS hst.groups_symbols_updated_at_idx;
DROP INDEX IF EXISTS hst.groups_symbols_group_path_idx;
DROP INDEX IF EXISTS hst.groups_symbols_group_id_idx;
DROP TABLE IF EXISTS hst.groups_symbols;
DROP SEQUENCE IF EXISTS hst.groups_symbols_symbol_id_seq;

DROP INDEX IF EXISTS hst.groups_updated_at_idx;
DROP INDEX IF EXISTS hst.groups_parent_id_idx;
DROP INDEX IF EXISTS hst.groups_group_uidx;
DROP TABLE IF EXISTS hst.groups;
DROP SEQUENCE IF EXISTS hst.groups_group_id_seq;

ALTER TABLE hst.clients DROP CONSTRAINT IF EXISTS clients_client_origin_login_fkey;
ALTER TABLE hst.clients DROP CONSTRAINT IF EXISTS clients_introducer_fkey;
DROP TABLE IF EXISTS hst.users CASCADE;
DROP SEQUENCE IF EXISTS hst.users_login_seq;

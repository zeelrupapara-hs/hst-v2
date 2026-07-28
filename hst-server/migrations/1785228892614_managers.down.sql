ALTER TABLE hst.clients DROP CONSTRAINT IF EXISTS clients_compliance_approved_by_fkey;
ALTER TABLE hst.clients DROP CONSTRAINT IF EXISTS clients_assigned_manager_fkey;
DROP TABLE IF EXISTS hst.managers CASCADE;

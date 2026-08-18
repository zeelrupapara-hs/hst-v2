DROP TABLE IF EXISTS hst.mail_attachments;
DROP INDEX IF EXISTS hst.mails_thread_idx;
ALTER TABLE hst.mails DROP COLUMN IF EXISTS thread_id, DROP COLUMN IF EXISTS attach_id;

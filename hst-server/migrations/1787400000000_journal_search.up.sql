-- a server-wide journal request is a LIKE over message across every login, and trade lines
-- from the engine will dominate the table, so the text gets a trigram index.
CREATE EXTENSION IF NOT EXISTS pg_trgm;
CREATE INDEX IF NOT EXISTS journal_message_trgm_idx ON hst.journal USING gin (message gin_trgm_ops);

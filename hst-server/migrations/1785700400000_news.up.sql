CREATE TABLE IF NOT EXISTS hst.news (
    news_id      BIGSERIAL PRIMARY KEY,
    external_id  TEXT   NOT NULL,
    title        TEXT   NOT NULL DEFAULT '',
    summary      TEXT   NOT NULL DEFAULT '',
    url          TEXT   NOT NULL DEFAULT '',
    source       TEXT   NOT NULL DEFAULT '',
    symbols      TEXT[] NOT NULL DEFAULT '{}',
    published_at BIGINT NOT NULL DEFAULT 0,
    created_at   BIGINT NOT NULL DEFAULT 0
);

CREATE UNIQUE INDEX IF NOT EXISTS news_external_id_uidx ON hst.news (external_id);
CREATE INDEX IF NOT EXISTS news_published_at_idx ON hst.news (published_at DESC);

COMMENT ON TABLE hst.news IS 'durable copy of the items hst-news ingests; redis holds only a 24h hot cache';
COMMENT ON COLUMN hst.news.external_id IS 'NewsItem.id from the feed, the dedup key';
COMMENT ON COLUMN hst.news.published_at IS 'unix nanoseconds';
COMMENT ON COLUMN hst.news.created_at IS 'unix nanoseconds';

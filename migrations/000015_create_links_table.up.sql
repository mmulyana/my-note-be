CREATE TABLE links (
    id          text PRIMARY KEY,
    note_id     text NOT NULL REFERENCES notes(id) ON DELETE CASCADE,
    url         text NOT NULL DEFAULT '',
    title       text NOT NULL DEFAULT '',
    description text NOT NULL DEFAULT '',
    image       text NOT NULL DEFAULT '',
    favicon     text NOT NULL DEFAULT '',
    site_name   text NOT NULL DEFAULT '',
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX links_note_id_idx ON links (note_id);
CREATE INDEX links_url_idx     ON links (url);

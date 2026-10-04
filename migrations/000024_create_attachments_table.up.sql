CREATE TABLE attachments (
    id         text PRIMARY KEY,
    note_id    text NOT NULL REFERENCES notes(id) ON DELETE CASCADE,
    path       text NOT NULL,
    thumb_path text NOT NULL DEFAULT '',
    mime       text NOT NULL DEFAULT '',
    size       bigint NOT NULL DEFAULT 0,
    width      integer NOT NULL DEFAULT 0,
    height     integer NOT NULL DEFAULT 0,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX attachments_note_id_idx ON attachments (note_id, created_at, id);

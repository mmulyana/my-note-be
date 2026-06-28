CREATE TYPE todo_priority AS ENUM ('low', 'medium', 'high');

CREATE TABLE notes (
    id          text PRIMARY KEY,
    user_id     uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title       varchar(200) NOT NULL DEFAULT '',
    preview     text NOT NULL DEFAULT '',
    content     text NOT NULL DEFAULT '',
    todo_total  int  NOT NULL DEFAULT 0,
    todo_done   int  NOT NULL DEFAULT 0,
    archived    boolean NOT NULL DEFAULT false,
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX notes_user_updated_idx ON notes (user_id, updated_at DESC);

CREATE TABLE todos (
    id          text PRIMARY KEY,
    note_id     text NOT NULL REFERENCES notes(id) ON DELETE CASCADE,
    text        text NOT NULL DEFAULT '',
    checked     boolean NOT NULL DEFAULT false,
    deadline    date,
    priority    todo_priority NOT NULL DEFAULT 'medium',
    tags        jsonb NOT NULL DEFAULT '[]',
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX todos_note_id_idx  ON todos (note_id);
CREATE INDEX todos_deadline_idx ON todos (deadline);
CREATE INDEX todos_checked_idx  ON todos (checked);

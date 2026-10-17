ALTER TABLE notes ADD COLUMN parent_id text NULL REFERENCES notes(id) ON DELETE SET NULL;

CREATE INDEX notes_parent_idx ON notes (parent_id) WHERE parent_id IS NOT NULL;

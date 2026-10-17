DROP INDEX IF EXISTS notes_parent_idx;
ALTER TABLE notes DROP COLUMN parent_id;

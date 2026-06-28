DROP INDEX IF EXISTS notes_text_search_idx;

ALTER TABLE notes DROP COLUMN IF EXISTS text;

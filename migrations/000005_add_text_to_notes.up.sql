ALTER TABLE notes ADD COLUMN text text NOT NULL DEFAULT '';

CREATE INDEX notes_text_search_idx ON notes USING gin(to_tsvector('simple', title || ' ' || text));

ALTER TABLE attachments ADD COLUMN is_cover boolean NOT NULL DEFAULT false;
CREATE UNIQUE INDEX attachments_one_cover_idx ON attachments (note_id) WHERE is_cover;

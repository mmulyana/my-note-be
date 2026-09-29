ALTER TABLE notes ADD COLUMN position double precision NOT NULL DEFAULT 0;
UPDATE notes SET position = -(EXTRACT(EPOCH FROM created_at) * 1000);
CREATE INDEX notes_user_position_idx ON notes (user_id, position, id);

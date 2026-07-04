ALTER TABLE notes ADD COLUMN pinned boolean NOT NULL DEFAULT false;
ALTER TABLE notes ADD COLUMN secret boolean NOT NULL DEFAULT false;

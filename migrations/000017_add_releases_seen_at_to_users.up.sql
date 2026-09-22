-- note: when the user last opened the updates panel; NULL means never, so
-- every published release counts as unread.
ALTER TABLE users ADD COLUMN releases_seen_at timestamptz;

DROP INDEX IF EXISTS idx_notes_folder_id;
ALTER TABLE notes DROP COLUMN IF EXISTS folder_id;

DROP INDEX IF EXISTS idx_folders_deleted_at;
DROP INDEX IF EXISTS idx_folders_user_id;
DROP TABLE IF EXISTS folders;

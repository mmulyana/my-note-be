-- Enforce unique folder name per user, ignoring soft-deleted rows so a name
-- can be reused after its folder is deleted.
CREATE UNIQUE INDEX idx_folders_user_name ON folders (user_id, name) WHERE deleted_at IS NULL;

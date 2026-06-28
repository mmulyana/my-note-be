CREATE TABLE folders (
    id         uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name       varchar(255) NOT NULL,
    color      varchar(32) NOT NULL DEFAULT 'default',
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    deleted_at timestamptz
);

CREATE INDEX idx_folders_user_id    ON folders (user_id);
CREATE INDEX idx_folders_deleted_at ON folders (deleted_at);

ALTER TABLE notes ADD COLUMN folder_id uuid NULL REFERENCES folders(id) ON DELETE SET NULL;
CREATE INDEX idx_notes_folder_id ON notes (folder_id);

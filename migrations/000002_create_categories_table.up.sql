CREATE TABLE categories (
    id      uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name    varchar(255) NOT NULL,
    UNIQUE (user_id, name)
);

CREATE INDEX idx_categories_user_id ON categories (user_id);

CREATE TABLE note_categories (
    note_id     text NOT NULL REFERENCES notes(id) ON DELETE CASCADE,
    category_id uuid NOT NULL REFERENCES categories(id) ON DELETE CASCADE,
    PRIMARY KEY (note_id, category_id)
);

CREATE INDEX idx_note_categories_category_id ON note_categories (category_id);

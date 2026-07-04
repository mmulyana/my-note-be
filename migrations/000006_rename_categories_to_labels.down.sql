-- Revert join table note_labels -> note_categories
ALTER INDEX idx_note_labels_label_id RENAME TO idx_note_categories_category_id;
ALTER TABLE note_labels RENAME COLUMN label_id TO category_id;
ALTER TABLE note_labels RENAME TO note_categories;

-- Revert labels -> categories
ALTER INDEX idx_labels_user_id RENAME TO idx_categories_user_id;
ALTER TABLE labels RENAME CONSTRAINT labels_user_id_name_key TO categories_user_id_name_key;
ALTER TABLE labels RENAME TO categories;

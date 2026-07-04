-- Rename categories -> labels
ALTER TABLE categories RENAME TO labels;
ALTER TABLE labels RENAME CONSTRAINT categories_user_id_name_key TO labels_user_id_name_key;
ALTER INDEX idx_categories_user_id RENAME TO idx_labels_user_id;

-- Rename join table note_categories -> note_labels
ALTER TABLE note_categories RENAME TO note_labels;
ALTER TABLE note_labels RENAME COLUMN category_id TO label_id;
ALTER INDEX idx_note_categories_category_id RENAME TO idx_note_labels_label_id;

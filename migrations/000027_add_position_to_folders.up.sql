ALTER TABLE folders ADD COLUMN position double precision NOT NULL DEFAULT 0;
UPDATE folders f SET position = r.rn FROM (
  SELECT id, ROW_NUMBER() OVER (PARTITION BY user_id ORDER BY name) AS rn FROM folders
) r WHERE f.id = r.id;
CREATE INDEX folders_user_position_idx ON folders (user_id, position, id);

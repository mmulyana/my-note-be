ALTER TABLE users ADD COLUMN is_guest boolean NOT NULL DEFAULT false;
ALTER TABLE users ALTER COLUMN email DROP NOT NULL;
ALTER TABLE users ALTER COLUMN password DROP NOT NULL;
ALTER TABLE users ALTER COLUMN created_at DROP DEFAULT;
ALTER TABLE users ADD CONSTRAINT users_guest_or_registered_chk
    CHECK (is_guest OR (email IS NOT NULL AND password IS NOT NULL));

CREATE TABLE ai_usage (
    user_id    uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    day        date NOT NULL,
    tokens     bigint NOT NULL DEFAULT 0,
    updated_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, day)
);

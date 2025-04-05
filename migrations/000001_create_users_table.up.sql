BEGIN;

CREATE SCHEMA IF NOT EXISTS cataloging;

CREATE TABLE IF NOT EXISTS "cataloging".users (
    id          TEXT        NOT NULL,
    name        TEXT        NOT NULL,
    email       TEXT        NOT NULL,
    password    TEXT        NOT NULL,
    is_admin    BOOLEAN     DEFAULT FALSE,
    created_at  TIMESTAMPTZ DEFAULT NOW(),
    updated_at  TIMESTAMPTZ DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ
);

CREATE UNIQUE INDEX IF NOT EXISTS users_id_unique ON "cataloging".users(id) WHERE deleted_at IS NULL;

COMMIT;
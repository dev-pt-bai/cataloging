BEGIN;

CREATE TABLE IF NOT EXISTS "cataloging".material_types (
    code        TEXT NOT NULL,
    description TEXT NOT NULL,
    val_class   TEXT,
    created_at  TIMESTAMPTZ DEFAULT NOW(),
    updated_at  TIMESTAMPTZ DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ
);

CREATE UNIQUE INDEX IF NOT EXISTS material_types_code_unique ON "cataloging".material_types(code) WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS "cataloging".material_uoms (
    code        TEXT NOT NULL,
    description TEXT NOT NULL,
    created_at  TIMESTAMPTZ DEFAULT NOW(),
    updated_at  TIMESTAMPTZ DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ
);

CREATE UNIQUE INDEX IF NOT EXISTS material_uoms_code_unique ON "cataloging".material_uoms(code) WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS "cataloging".material_groups (
    code        TEXT NOT NULL,
    description TEXT NOT NULL,
    created_at  TIMESTAMPTZ DEFAULT NOW(),
    updated_at  TIMESTAMPTZ DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ
);

CREATE UNIQUE INDEX IF NOT EXISTS material_groups_code_unique ON "cataloging".material_groups(code) WHERE deleted_at IS NULL;

COMMIT;
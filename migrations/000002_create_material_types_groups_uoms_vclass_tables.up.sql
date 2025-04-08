SET autocommit = OFF;

BEGIN;

CREATE TABLE IF NOT EXISTS material_types (
    code        VARCHAR(255)  NOT NULL,
    description VARCHAR(1023) NOT NULL,
    val_class   VARCHAR(255),
    created_at  INT UNSIGNED  DEFAULT (UNIX_TIMESTAMP()),
    updated_at  INT UNSIGNED  DEFAULT (UNIX_TIMESTAMP()),
    deleted_at  INT UNSIGNED  DEFAULT 0,

    PRIMARY KEY (code, deleted_at)
);

CREATE TABLE IF NOT EXISTS material_uoms (
    code        VARCHAR(255)  NOT NULL,
    description VARCHAR(1023) NOT NULL,
    created_at  INT UNSIGNED  DEFAULT (UNIX_TIMESTAMP()),
    updated_at  INT UNSIGNED  DEFAULT (UNIX_TIMESTAMP()),
    deleted_at  INT UNSIGNED  DEFAULT 0,

    PRIMARY KEY (code, deleted_at)
);

CREATE TABLE IF NOT EXISTS material_groups (
    code        VARCHAR(255)  NOT NULL,
    description VARCHAR(1023) NOT NULL,
    created_at  INT UNSIGNED  DEFAULT (UNIX_TIMESTAMP()),
    updated_at  INT UNSIGNED  DEFAULT (UNIX_TIMESTAMP()),
    deleted_at  INT UNSIGNED  DEFAULT 0,

    PRIMARY KEY (code, deleted_at)
);

COMMIT;

SET autocommit = ON;
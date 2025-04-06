SET autocommit = OFF;

BEGIN;

CREATE TABLE IF NOT EXISTS users (
    id          VARCHAR(255)    NOT NULL,
    name        VARCHAR(255)    NOT NULL,
    email       VARCHAR(255)    NOT NULL,
    password    VARCHAR(255)    NOT NULL,
    is_admin    TINYINT(1)      DEFAULT 0,
    created_at  INT UNSIGNED    DEFAULT (UNIX_TIMESTAMP()),
    updated_at  INT UNSIGNED    DEFAULT (UNIX_TIMESTAMP()),
    deleted_at  INT UNSIGNED    DEFAULT 0,

    PRIMARY KEY (id, deleted_at)
);

COMMIT;

SET autocommit = ON;
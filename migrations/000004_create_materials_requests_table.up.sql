SET autocommit = OFF;

BEGIN;

CREATE TABLE IF NOT EXISTS requests (
    id           VARCHAR(255) NOT NULL,
    subject      VARCHAR(255) NOT NULL,
    is_new       TINYINT(1)   NOT NULL,
    requested_by VARCHAR(255) NOT NULL,
    status       TINYINT      NOT NULL,
    created_at   INT UNSIGNED DEFAULT (UNIX_TIMESTAMP()),
    updated_at   INT UNSIGNED DEFAULT (UNIX_TIMESTAMP()),
    deleted_at   INT UNSIGNED DEFAULT 0,

    PRIMARY KEY (id)
);

CREATE TABLE IF NOT EXISTS plants (
    code        VARCHAR(255)  NOT NULL,
    description VARCHAR(1023) NOT NULL,
    created_at  INT UNSIGNED  DEFAULT (UNIX_TIMESTAMP()),
    updated_at  INT UNSIGNED  DEFAULT (UNIX_TIMESTAMP()),
    deleted_at  INT UNSIGNED  DEFAULT 0,

    PRIMARY KEY (code, deleted_at)
);

CREATE TABLE IF NOT EXISTS manufacturers (
    code        VARCHAR(255)  NOT NULL,
    description VARCHAR(1023) NOT NULL,
    created_at  INT UNSIGNED  DEFAULT (UNIX_TIMESTAMP()),
    updated_at  INT UNSIGNED  DEFAULT (UNIX_TIMESTAMP()),
    deleted_at  INT UNSIGNED  DEFAULT 0,

    PRIMARY KEY (code, deleted_at)
);

CREATE TABLE IF NOT EXISTS materials (
    id                VARCHAR(255)  NOT NULL,
    number            VARCHAR(255),
    plant_code        VARCHAR(255)  NOT NULL,
    type_code         VARCHAR(255)  NOT NULL,
    uom_code          VARCHAR(255)  NOT NULL,
    group_code        VARCHAR(255)  NOT NULL,
    equipment_code    VARCHAR(255),
    manufacturer_code VARCHAR(255),
    short_text        VARCHAR(255),
    long_text         VARCHAR(1023) NOT NULL,
    note              VARCHAR(1023),
    status            TINYINT       NOT NULL,
    request_id        VARCHAR(255)  NOT NULL,
    created_at        INT UNSIGNED  DEFAULT (UNIX_TIMESTAMP()),
    updated_at        INT UNSIGNED  DEFAULT (UNIX_TIMESTAMP()),
    deleted_at        INT UNSIGNED  DEFAULT 0,

    PRIMARY KEY (id),
    FOREIGN KEY (request_id) REFERENCES requests (id)
);

CREATE TABLE IF NOT EXISTS assets (
    id           VARCHAR(255)  NOT NULL,
    name         VARCHAR(255)  NOT NULL,
    size         INT UNSIGNED  NOT NULL,
    download_url VARCHAR(2047) NOT NULL,
    web_url      VARCHAR(2047) NOT NULL,
    created_by   VARCHAR(255)  NOT NULL,
    material_id  VARCHAR(255),
    created_at   INT UNSIGNED  DEFAULT (UNIX_TIMESTAMP()),
    updated_at   INT UNSIGNED  DEFAULT (UNIX_TIMESTAMP()),
    deleted_at   INT UNSIGNED  DEFAULT 0,

    PRIMARY KEY (id)
);

CREATE INDEX asset_material_idx ON assets (material_id); 

COMMIT;

SET autocommit = ON;
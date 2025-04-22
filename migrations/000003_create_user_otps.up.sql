SET autocommit = OFF;

BEGIN;

CREATE TABLE IF NOT EXISTS user_otps (
    user_id     VARCHAR(255)    NOT NULL,
    user_email  VARCHAR(255)    NOT NULL,
    otp         VARCHAR(255)    NOT NULL,
    created_at  INT UNSIGNED    DEFAULT (UNIX_TIMESTAMP()),
    expired_at  INT UNSIGNED    NOT NULL,

    PRIMARY KEY (user_id, otp)
);

COMMIT;

SET autocommit = ON;
SET autocommit = OFF;

BEGIN;

DROP TABLE IF EXISTS assets;

DROP TABLE IF EXISTS materials;

DROP TABLE IF EXISTS plants;

DROP TABLE IF EXISTS requests;

COMMIT;

SET autocommit = ON;
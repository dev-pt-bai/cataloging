SET autocommit = OFF;

BEGIN;

DROP TABLE IF EXISTS material_groups;

DROP TABLE IF EXISTS material_uoms;

DROP TABLE IF EXISTS material_types;

COMMIT;

SET autocommit = ON;
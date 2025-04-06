# Database Migrations Guideline

## Prerequisites

Before you begin, ensure you have the following installed:

- MySQL

### Optional migration tools

Although you can create, deploy and revert the migration scripts manually, we encourage you to use [go-migrate](https://github.com/golang-migrate/migrate/tree/master/cmd/migrate), an easy-to-use Go-based CLI application to handle database migrations. It is available on Windows, MacOS or Linux-based operating systems.

## Creating database

If you need to create a new database, run this script:

```sql
CREATE DATABASE bai;
```

It will create a new database named `bai`.

## Adding new migrations scripts

All migratons scripts should be placed in `migrations` directory.

There should always be two kind of scripts, i.e. deploy and revert. As the name suggests, a deploy script will deploy the migration script, while a revert script will do the opposite.

Both the deploy and revert scripts should follow a sequence. The deploy scripts should be executed forward from the beginning to the most recent, while the revert scripts should be executed backward from the most recent to the desired stage. To organize the sequence, we use numbers as part of the script's name, e.g.

```
000001_create_users_table.up.sql
000001_create_users_table.down.sql
```

The `up` and `down` infix denote deploy and revert scripts, respectively.

The script's name should be made concise, yet as descriptive as possible. In the above example, it is clear that the script is about creating a new table named users.

The next deploy and revert scripts should be numbered successively. For example,

```
000002_create_material_table.up.sql
000002_create_material_table.down.sql
```

If you use go-migrate, just run this command (at the root of project's directory)

```bash
migrate create -ext sql -dir migrations -seq create_users_table
```

for the first set of deploy and revert scripts, and

```bash
migrate create -ext sql -dir migrations -seq create_material_table
```

for the second set of deploy and revert scripts. go-migrate will automatically organize the sequence for you and place the associated empty files in the correct directory.

In writing the deploy and revert scripts, always wrap it in a transaction to maintain the integrity of the database.

For example, write:

```sql
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
```

instead of

```sql
CREATE TABLE IF NOT EXISTS users (
    id          VARCHAR(255)    NOT NULL,
    name        VARCHAR(255)    NOT NULL,
    email       VARCHAR(255)    NOT NULL,
    password    VARCHAR(255)    NOT NULL,
    is_admin    TINYINT(1)      DEFAULT 0,
    created_at  INT UNSIGNED    DEFAULT (UNIX_TIMESTAMP()),
    updated_at  INT UNSIGNED    DEFAULT (UNIX_TIMESTAMP()),
    deleted_at  INT UNSIGNED    DEFAULT 0
);

CREATE UNIQUE INDEX unique_id_idx ON users (id, deleted_at);
```

## Deploying the migrations

You may manually run the scripts sequentially forward one by one in the database directly.

If you use go-migrate, all it takes is just run this command:

```bash
migrate -database "mysql://[user][:password]@tcp([host][:port])/[database_name]" -path migrations up
```

## Reverting the migrations

Similar to the deploying steps, you may manually run the scripts sequentially backward one by one in the database directly, or use go-migrate with this command:

```bash
migrate -database "mysql://[user][:password]@tcp([host][:port])/[database_name]" -path migrations down
```

The above command will revert **ALL** migrations. To revert partially, use this command instead:

```bash
migrate -database "mysql://[user][:password]@tcp([host][:port])/[database_name]" -path migrations down [step]
```

Just specify `step` to tell go-migrate, how many migration scripts will be rolled back from the most recent.
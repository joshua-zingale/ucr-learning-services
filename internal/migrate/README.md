A flexible migration utility.

`migrate` is designed with database migrations in mind, but it could be used to
manage versioning of any system that uses files to transition from one state to
the next.

## Usage

`migrate` takes in a folder as input that contains migration files that contain
either `.down.` or `.up` in their name. These files are internally sorted
lexicographically, where a file being lexicographically smaller are earlier
migrations and those being lexicographical larger are later migrations. The
files with names containing `.up` should upgrade a system while those with names
containing `.down` downgrade the system.

For example, we may have the following directory of db migrations

```
migrations
├── 1.down.sql
├── 1.up.sql
├── 2.down.sql
├── 2.up.sql
├── 3.down.sql
├── 3.up.sql
├── 4.down.sql
├── 4.up.sql
├── 5.down.sql
├── 5.up.sql
├── 6.down.sql
├── 6.up.sql
├── 7.down.sql
├── 7.up.sql
├── 8.down.sql
├── 8.up.sql
├── 9.down.sql
└── 9.up.sql
```

Then, assuming that none of the migrations have been run, running the following
would run execute each migration as a database transaction with `psql` up to and
including `5.up.sql`

```bash
migrate -to 5 -command 'psql -U db_user -d db_name -f "@#" -1' migrations
```

`@#` is a special character sequence that will be swapped for the name of the
migration. Thus will this execute the following in order, stopping should any
fail but not attempting to revert automatically:

```
psql -U db_user -d db_name -f "migrations/1.up.sql" -1
psql -U db_user -d db_name -f "migrations/2.up.sql" -1
psql -U db_user -d db_name -f "migrations/3.up.sql" -1
psql -U db_user -d db_name -f "migrations/4.up.sql" -1
psql -U db_user -d db_name -f "migrations/5.up.sql" -1
```

Then, we could revert to version 3 by doing

```bash
migrate -to 3 -command 'psql -U db_user -d db_name -f "@#" -1' migrations
```

which will execute

```
psql -U db_user -d db_name -f "migrations/5.down.sql" -1
psql -U db_user -d db_name -f "migrations/4.down.sql" -1
```

`migrate` remembers the migration state by storing a `.migrationstate` file
inside the specified migrations folder.

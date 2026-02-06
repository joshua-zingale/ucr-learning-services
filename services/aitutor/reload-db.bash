set -e

db_to_delete=${1:-aitutor_go}

sqlc generate

dropdb -f --if-exists $db_to_delete
psql -c "create database $db_to_delete;"
psql -d $db_to_delete -f db/schema.sql
psql -d $db_to_delete -f db/mocks/0001.sql
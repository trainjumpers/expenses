# Expenses

## Setting up local Postgres
Setup local postgres server and create user and database for local development. This step has to be done only once, for fresh devs
```bash
sudo service postgresql start # start the postgresql service
sudo -i -u postgres # postgres shell login as postgres user
createuser testuser  # create new user 'testuser'
createdb testdb -O testuser  # create new db using user 'testuser'
psql testdb  # start postgres shell and use database 'testdb'
alter user testuser with password 'test';  # attach password 'test' to user 'testuser'
create schema test;  # create new schema on testdb
grant all privileges on database testdb to testuser;
grant all privileges on schema test to testuser;
```

## Delete and Recreate tables
Once the database is created, create all the tables by running the `delete_and_recreate_tables.sh` script
```bash
./delete_and_recreate_tables.sh
```
Windows users can run this script using the git bash terminal

## How to run Server
Once the database is setup, its time to run the server
```bash
go get . # Install Project Dependencies
go run . # Build and run the server
```

For Live reloading, install [air](https://github.com/cosmtrek/air), and run
```bash
air
```
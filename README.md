# Expenses

## Setting up local Postgres
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

## How to run Server
```bash
go get . # Install Project Dependencies
go run . # Build and run the server
```

For Live reloading, install [air](https://github.com/cosmtrek/air), and run
```bash
air
```
# Expenses Management
A simple expenses management application built with Go, React, and Postgres. It allows users to track their expenses, categorize them, and generate reports. It is specifically made for Indians with an option to parse a CSV file of a bank statement from Axis, SBI, or HDFC bank.

## Project Command Automation with `just`

This project utilizes [Just](https://github.com/casey/just), a command runner designed to simplify and automate common tasks. The `justfile` defines a set of recipes (commands) to streamline development workflows, including database migrations, server management, and dependency installation.

### Prerequisites

Ensure you have set up variables as defined in the `.env.example` file.

### Available Recipes
Use this command to view all available commands
```bash
just --list
```

## How to Use

To execute a recipe, run the following command:

```bash
just <recipe-name> [arguments]
```

Run this to install required dependencies
```bash
just install
```

## Setting up server

### Setting up local Postgres
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

### How to run Server
Once the database is setup, it's time to run the server. This will install air for live reloading
```bash
just run # Build and run the server
```

## Migrations

This project uses [Goose](https://github.com/pressly/goose) for database migrations. Goose is a database migration tool that helps manage schema changes in a structured and version-controlled way. It is recommended to invoke it using justfile. 

For example to upgrade the database schema to the latest version, run:
```bash
just db-upgrade
```

For more commands, run 
```bash
just -l
```
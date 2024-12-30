#!/bin/bash

# Database credentials
export $(grep -v '^#' .env | xargs)

DB_USER=$PGUSER
DB_PASS=$PGPASSWORD
DB_NAME=$PGDBNAME
DB_HOST=$PGHOST
DB_SCHEMA=$PGSCHEMA

# Connect to the PostgreSQL database and delete all tables in the schema
echo "Deleting all tables in the $PGSCHEMA schema..."
PGPASSWORD=$DB_PASS psql -U $DB_USER -h $DB_HOST -d $DB_NAME -c "
DO
\$do\$
DECLARE
   _tbl text;
BEGIN
   FOR _tbl  IN (SELECT tablename FROM pg_tables WHERE schemaname = '$PGSCHEMA')
   LOOP
      EXECUTE 'DROP TABLE IF EXISTS $PGSCHEMA.' || _tbl || ' CASCADE';
   END LOOP;
END
\$do\$;
"
# Create user table
echo "Creating 'user' table..."
PGPASSWORD=$DB_PASS psql -U $DB_USER -h $DB_HOST -d $DB_NAME -c "
   CREATE TABLE $PGSCHEMA.user (
      id SERIAL PRIMARY KEY,
      name VARCHAR(100) NOT NULL,
      email VARCHAR(100) UNIQUE NOT NULL,
      password VARCHAR(100) NOT NULL,
      deleted_at TIMESTAMPTZ NULL
   );
"

# Create expense table
# Unique ID stores a precalculated unique id for each expense based on
# date, payer id, expense id, name and desc. 
# This is used to prevent duplicate expense while batch processing
# Also Note: Negative amount indicates a credit transaction
# Positive amount indicates a debit transaction
echo "Creating 'expense' table..."
PGPASSWORD=$DB_PASS psql -U $DB_USER -h $DB_HOST -d $DB_NAME -c "
   CREATE TABLE $PGSCHEMA.expense (
      id SERIAL PRIMARY KEY,
      amount FLOAT NOT NULL,
      payer_id INTEGER NOT NULL,
      name VARCHAR(100) NULL,
      description TEXT NULL,
      created_by INTEGER NOT NULL,
      unique_id VARCHAR(255) UNIQUE NOT NULL,
      created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
      CONSTRAINT fk_user FOREIGN KEY (payer_id) REFERENCES $PGSCHEMA.user(id),
      CONSTRAINT fk_created_by FOREIGN KEY (created_by) REFERENCES $PGSCHEMA.user(id)
   );
"

# Create expense_contributor table
echo "Creating 'expense_user_mapping' table..."
PGPASSWORD=$DB_PASS psql -U $DB_USER -h $DB_HOST -d $DB_NAME -c "
   CREATE TABLE $PGSCHEMA.expense_user_mapping (
      id SERIAL PRIMARY KEY,
      expense_id INTEGER NOT NULL,
      user_id INTEGER NOT NULL,
      amount FLOAT NOT NULL,
      CONSTRAINT fk_expense FOREIGN KEY (expense_id) REFERENCES $PGSCHEMA.expense(id),
      CONSTRAINT fk_user FOREIGN KEY (user_id) REFERENCES $PGSCHEMA.user(id)
   );
"

# Create a job table
echo "Creating 'job' table..."
PGPASSWORD=$DB_PASS psql -U $DB_USER -h $DB_HOST -d $DB_NAME -c "
   CREATE TABLE $PGSCHEMA.jobs (
      id SERIAL PRIMARY KEY,
      name VARCHAR(100) NOT NULL,
      status VARCHAR(100) NOT NULL DEFAULT 'created',
      metadata JSONB NULL,
      created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
      updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
   );
"

# Create job_user_mapping table
echo "Creating 'jobs_user_mapping' table..."
PGPASSWORD=$DB_PASS psql -U $DB_USER -h $DB_HOST -d $DB_NAME -c "
   CREATE TABLE $PGSCHEMA.job_user_mapping (
      id SERIAL PRIMARY KEY,
      job_id INTEGER NOT NULL,
      user_id INTEGER NOT NULL,
      CONSTRAINT fk_job FOREIGN KEY (job_id) REFERENCES $PGSCHEMA.jobs(id),
      CONSTRAINT fk_user FOREIGN KEY (user_id) REFERENCES $PGSCHEMA.user(id)
   );
"

# Create categories table
echo "Creating 'categories' table..."
PGPASSWORD=$DB_PASS psql -U $DB_USER -h $DB_HOST -d $DB_NAME -c "
   CREATE TABLE $PGSCHEMA.categories (
      id SERIAL PRIMARY KEY,
      name VARCHAR(100) NOT NULL,
      color VARCHAR(100) NULL,
      created_by INTEGER NOT NULL,
      created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
      CONSTRAINT fk_created_by FOREIGN KEY (created_by) REFERENCES $PGSCHEMA.user(id)
   );
"

# Create subcategories table
echo "Creating 'subcategories' table..."
PGPASSWORD=$DB_PASS psql -U $DB_USER -h $DB_HOST -d $DB_NAME -c "
   CREATE TABLE $PGSCHEMA.subcategories (
      id SERIAL PRIMARY KEY,
      name VARCHAR(100) NOT NULL,
      color VARCHAR(100) NULL,
      created_by INTEGER NOT NULL,
      created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
      CONSTRAINT fk_created_by FOREIGN KEY (created_by) REFERENCES $PGSCHEMA.user(id)
   );
"

# Create category_subcategory_mapping table
echo "Creating 'category_subcategory_mapping' table..."
PGPASSWORD=$DB_PASS psql -U $DB_USER -h $DB_HOST -d $DB_NAME -c "
   CREATE TABLE $PGSCHEMA.category_subcategory_mapping (
      id SERIAL PRIMARY KEY,
      category_id INTEGER NOT NULL,
      subcategory_id INTEGER NOT NULL,
      CONSTRAINT fk_category FOREIGN KEY (category_id) REFERENCES $PGSCHEMA.categories(id),
      CONSTRAINT fk_subcategory FOREIGN KEY (subcategory_id) REFERENCES $PGSCHEMA.subcategories(id)
   );
"

# Create subcategory rules table
echo "Creating 'subcategory_rules' table..."
PGPASSWORD=$DB_PASS psql -U $DB_USER -h $DB_HOST -d $DB_NAME -c "
   CREATE TABLE $PGSCHEMA.subcategory_rules (
      id SERIAL PRIMARY KEY,
      rule JSONB NOT NULL,
      subcategory_id INTEGER NOT NULL,
      CONSTRAINT fk_subcategory FOREIGN KEY (subcategory_id) REFERENCES $PGSCHEMA.subcategories(id)
   );
"
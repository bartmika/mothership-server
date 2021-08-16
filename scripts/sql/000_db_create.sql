drop database mothership_db;
create database mothership_db;
\c mothership_db;
CREATE USER golang WITH PASSWORD '123password';
GRANT ALL PRIVILEGES ON DATABASE mothership_db to golang;
ALTER USER golang CREATEDB;
ALTER ROLE golang SUPERUSER;

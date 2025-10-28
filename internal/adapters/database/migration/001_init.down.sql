-- This file is for rolling back the schema changes.
-- Drop tables in reverse order of creation to respect foreign key constraints

DROP TABLE IF EXISTS tokens CASCADE;
DROP TABLE IF EXISTS credentials CASCADE;
DROP TABLE IF EXISTS user_roles CASCADE;
DROP TABLE IF EXISTS users CASCADE;
DROP TABLE IF EXISTS roles CASCADE;
DROP TABLE IF EXISTS designations CASCADE;
DROP TABLE IF EXISTS departments CASCADE;

DROP TYPE IF EXISTS token_scope;
DROP TYPE IF EXISTS user_status;

DROP EXTENSION IF EXISTS "uuid-ossp";

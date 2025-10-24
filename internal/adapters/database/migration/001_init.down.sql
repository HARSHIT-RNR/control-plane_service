-- This file is for rolling back the schema changes.

DROP INDEX IF EXISTS idx_user_tenant_roles_role_id;
DROP INDEX IF EXISTS idx_user_tenant_roles_user_id;
DROP INDEX IF EXISTS idx_users_email;

ALTER TABLE users DROP CONSTRAINT IF EXISTS fk_designation;
ALTER TABLE users DROP CONSTRAINT IF EXISTS fk_department;

DROP TABLE IF EXISTS designations;
DROP TABLE IF EXISTS departments;
DROP TABLE IF EXISTS user_tenant_roles;
DROP TABLE IF EXISTS roles;
DROP TABLE IF EXISTS users;

DROP TYPE IF EXISTS user_status;

DROP EXTENSION IF EXISTS "uuid-ossp";
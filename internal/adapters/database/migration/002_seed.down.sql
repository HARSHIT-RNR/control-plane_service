-- This file removes the seed data added in the 'up' migration.
-- It's useful for resetting the database to a clean state without dropping the tables.

-- Delete the specific seed data for the test tenant.
-- Replace 'tenant-123' with the same tenant ID used in the 'up' file.

DELETE FROM designations WHERE tenant_id = 'tenant-123' AND name IN
    ('Line Operator', 'QA Inspector', 'Forklift Operator', 'HR Manager');

DELETE FROM departments WHERE tenant_id = 'tenant-123' AND name IN
    ('Manufacturing', 'Quality Assurance', 'Logistics', 'Human Resources');

DELETE FROM roles WHERE tenant_id = 'tenant-123' AND name IN
    ('Admin', 'Member', 'Viewer');


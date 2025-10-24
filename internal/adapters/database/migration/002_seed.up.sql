-- This file seeds the database with essential data for development and testing.
-- Replace 'tenant-123' throughout this file with a tenant ID you use for local testing.

-- Seed default roles for the test tenant.
INSERT INTO roles (tenant_id, name, description, permissions)
VALUES
    ('tenant-123', 'Admin', 'Full administrative access to the tenant.', ARRAY['users:create', 'users:read', 'users:update', 'users:delete', 'roles:manage', 'departments:manage', 'designations:manage']),
    ('tenant-123', 'Member', 'Standard user with access to core features.', ARRAY['invoices:create', 'invoices:read', 'customers:read']),
    ('tenant-123', 'Viewer', 'Read-only access to most resources.', ARRAY['invoices:read', 'customers:read'])
ON CONFLICT (tenant_id, name) DO NOTHING;

-- Seed default departments for the test tenant.
INSERT INTO departments (tenant_id, name, description)
VALUES
    ('tenant-123', 'Manufacturing', 'Floor operations and production line.'),
    ('tenant-123', 'Quality Assurance', 'Product testing and quality control.'),
    ('tenant-123', 'Logistics', 'Shipping, receiving, and warehouse management.'),
    ('tenant-123', 'Human Resources', 'Employee management and payroll.')
ON CONFLICT (tenant_id, name) DO NOTHING;

-- Seed default designations for the test tenant.
INSERT INTO designations (tenant_id, name, description)
VALUES
    ('tenant-123', 'Line Operator', 'Operates machinery on the production line.'),
    ('tenant-123', 'QA Inspector', 'Inspects products for quality defects.'),
    ('tenant-123', 'Forklift Operator', 'Moves materials within the warehouse.'),
    ('tenant-123', 'HR Manager', 'Manages the Human Resources department.')
ON CONFLICT (tenant_id, name) DO NOTHING;


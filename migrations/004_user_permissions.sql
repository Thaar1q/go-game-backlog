INSERT INTO permissions (name, description) VALUES
    ('game:list',    'View catalog games'),
    ('game:create',  'Add new game to catalog'),
    ('game:update',  'Edit existing catalog game'),
    ('game:delete',  'Delete catalog game'),
    ('role:assign',  'Assign roles to users'),
    ('backlog:read:any',   'View any user backlog'),
    ('backlog:update:any', 'Update any user backlog'),
    ('backlog:delete:any', 'Delete any user backlog')
ON CONFLICT (name) DO NOTHING;

-- Admin: all permissions
INSERT INTO role_permissions (role_name, permission_name) VALUES
    ('admin', 'game:list'),
    ('admin', 'game:create'),
    ('admin', 'game:update'),
    ('admin', 'game:delete'),
    ('admin', 'role:assign'),
    ('admin', 'backlog:read:any'),
    ('admin', 'backlog:update:any'),
    ('admin', 'backlog:delete:any')
ON CONFLICT DO NOTHING;

-- Curator: catalog management
INSERT INTO role_permissions (role_name, permission_name) VALUES
    ('curator', 'game:list'),
    ('curator', 'game:create'),
    ('curator', 'game:update')
ON CONFLICT DO NOTHING;

-- user: read catalog
INSERT INTO role_permissions (role_name, permission_name) VALUES
    ('user', 'game:list')
ON CONFLICT DO NOTHING;

-- Bind FK
UPDATE users SET role = 'user' WHERE role NOT IN (SELECT name FROM roles);

ALTER TABLE users DROP CONSTRAINT IF EXISTS fkey_users_role;
ALTER TABLE users
    ADD CONSTRAINT fkey_users_role
    FOREIGN KEY (role) REFERENCES roles(name) ON UPDATE CASCADE;

CREATE INDEX IF NOT EXISTS idx_users_role ON users (role);

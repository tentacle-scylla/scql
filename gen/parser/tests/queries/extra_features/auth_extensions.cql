-- Extended Authentication/Authorization features

-- Create role with NOLOGIN (service accounts)
CREATE ROLE service_account WITH NOLOGIN = true;
CREATE ROLE batch_processor WITH PASSWORD = 'secret' AND NOLOGIN = false;

-- Create role with all options
CREATE ROLE admin_user
    WITH PASSWORD = 'secret123'
    AND LOGIN = true
    AND SUPERUSER = true;

CREATE ROLE IF NOT EXISTS readonly_user
    WITH PASSWORD = 'readonly'
    AND LOGIN = true
    AND SUPERUSER = false;

-- Alter role NOLOGIN
ALTER ROLE service_account WITH NOLOGIN = false;
ALTER ROLE batch_processor WITH LOGIN = true;

-- List users (alternative syntax)
LIST USERS;

-- List roles with details
LIST ROLES;
LIST ROLES OF admin_user;
LIST ALL ROLES OF admin_user NORECURSIVE;

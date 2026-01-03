-- Role management

CREATE ROLE admin;
CREATE ROLE IF NOT EXISTS admin;
CREATE ROLE admin WITH PASSWORD = 'secret' AND LOGIN = true;
CREATE ROLE admin WITH PASSWORD = 'secret' AND LOGIN = true AND SUPERUSER = true;
ALTER ROLE admin WITH PASSWORD = 'newsecret';
DROP ROLE admin;
DROP ROLE IF EXISTS admin;
LIST ROLES;
LIST ROLES OF admin;

-- Permission management

GRANT SELECT ON keyspace1.users TO admin;
GRANT ALL PERMISSIONS ON KEYSPACE keyspace1 TO admin;
GRANT MODIFY ON ALL KEYSPACES TO admin;
REVOKE SELECT ON keyspace1.users FROM admin;
REVOKE ALL PERMISSIONS ON KEYSPACE keyspace1 FROM admin;
LIST ALL PERMISSIONS;
LIST ALL PERMISSIONS OF admin;
LIST SELECT ON keyspace1.users;

-- NOTE: GRANT/REVOKE role TO/FROM role syntax not yet supported
-- GRANT admin TO developer;
-- REVOKE admin FROM developer;

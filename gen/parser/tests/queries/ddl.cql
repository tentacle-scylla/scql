-- Keyspace operations

CREATE KEYSPACE test WITH replication = {'class': 'SimpleStrategy', 'replication_factor': 1};
CREATE KEYSPACE IF NOT EXISTS test WITH replication = {'class': 'NetworkTopologyStrategy', 'dc1': 3, 'dc2': 2};
ALTER KEYSPACE test WITH replication = {'class': 'SimpleStrategy', 'replication_factor': 3};
DROP KEYSPACE test;
DROP KEYSPACE IF EXISTS test;

-- Table operations

CREATE TABLE users (id int PRIMARY KEY, name text, email text);
CREATE TABLE users (id int, name text, email text, PRIMARY KEY (id));
CREATE TABLE users (id int, cluster_id int, name text, PRIMARY KEY (id, cluster_id));
CREATE TABLE users (id int, cluster_id int, name text, PRIMARY KEY ((id), cluster_id));
CREATE TABLE users (id int, cluster_id int, ts timestamp, name text, PRIMARY KEY ((id, cluster_id), ts)) WITH CLUSTERING ORDER BY (ts DESC);
CREATE TABLE IF NOT EXISTS users (id int PRIMARY KEY, name text);
CREATE TABLE users (id int PRIMARY KEY, name text) WITH comment = 'Users table' AND gc_grace_seconds = 86400;
CREATE TABLE users (id int PRIMARY KEY, name text) WITH compaction = {'class': 'LeveledCompactionStrategy'};

ALTER TABLE users ADD age int;
ALTER TABLE users DROP email;
ALTER TABLE users RENAME id TO user_id;
ALTER TABLE users WITH gc_grace_seconds = 172800;

DROP TABLE users;
DROP TABLE IF EXISTS users;
TRUNCATE users;
TRUNCATE TABLE users;

-- Index operations

CREATE INDEX ON users (name);
CREATE INDEX user_name_idx ON users (name);
CREATE INDEX IF NOT EXISTS user_name_idx ON users (name);
CREATE INDEX ON users (KEYS(data));
-- NOTE: VALUES() index syntax not yet supported
-- CREATE INDEX ON users (VALUES(data));
CREATE INDEX ON users (ENTRIES(data));
CREATE INDEX ON users (FULL(data));
DROP INDEX user_name_idx;
DROP INDEX IF EXISTS user_name_idx;

-- Materialized views

-- NOTE: SELECT * in CREATE MATERIALIZED VIEW not yet supported, use explicit columns
-- CREATE MATERIALIZED VIEW users_by_name AS SELECT * FROM users WHERE name IS NOT NULL AND id IS NOT NULL PRIMARY KEY (name, id);
CREATE MATERIALIZED VIEW users_by_name AS SELECT id, name FROM users WHERE name IS NOT NULL AND id IS NOT NULL PRIMARY KEY (name, id);
DROP MATERIALIZED VIEW users_by_name;
DROP MATERIALIZED VIEW IF EXISTS users_by_name;

-- User-defined types

CREATE TYPE address (street text, city text, zip text);
CREATE TYPE IF NOT EXISTS address (street text, city text, zip text);
ALTER TYPE address ADD country text;
ALTER TYPE address RENAME zip TO postal_code;
DROP TYPE address;
DROP TYPE IF EXISTS address;

-- User-defined functions

CREATE FUNCTION my_func(input int) RETURNS NULL ON NULL INPUT RETURNS int LANGUAGE java AS 'return input * 2;';
CREATE OR REPLACE FUNCTION my_func(input int) CALLED ON NULL INPUT RETURNS int LANGUAGE java AS 'return input;';
DROP FUNCTION my_func;
DROP FUNCTION IF EXISTS my_func;

-- User-defined aggregates

-- NOTE: CREATE AGGREGATE requires FINALFUNC and INITCOND in base grammar
-- A simpler CREATE AGGREGATE syntax is not supported
-- CREATE AGGREGATE my_agg(int) SFUNC my_state_func STYPE int;
DROP AGGREGATE my_agg;
DROP AGGREGATE IF EXISTS my_agg;

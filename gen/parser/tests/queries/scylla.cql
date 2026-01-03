-- ScyllaDB-specific features

-- BYPASS CACHE
SELECT * FROM users WHERE id = 1 BYPASS CACHE;

-- USING TIMEOUT
SELECT * FROM users WHERE id = 1 USING TIMEOUT 5000;
INSERT INTO users (id, name) VALUES (1, 'Alice') USING TIMEOUT 5000;
UPDATE users USING TIMEOUT 5000 SET name = 'Bob' WHERE id = 1;
DELETE FROM users USING TIMEOUT 5000 WHERE id = 1;

-- PER PARTITION LIMIT
SELECT * FROM users PER PARTITION LIMIT 10;
SELECT * FROM users WHERE id = 1 PER PARTITION LIMIT 5 LIMIT 100;

-- PRUNE MATERIALIZED VIEW (requires ScyllaDB 5.0+)
-- PRUNE MATERIALIZED VIEW users_by_name;

-- Service levels (ScyllaDB Enterprise)
-- CREATE SERVICE LEVEL sl1 WITH timeout = 500ms AND workload_type = 'interactive';
-- ALTER SERVICE LEVEL sl1 WITH timeout = 1s;
-- DROP SERVICE LEVEL sl1;
-- ATTACH SERVICE LEVEL sl1 TO admin;
-- DETACH SERVICE LEVEL FROM admin;
-- LIST ALL SERVICE LEVELS;
-- LIST SERVICE LEVEL sl1;
-- LIST ATTACHED SERVICE LEVEL OF admin;
-- LIST EFFECTIVE SERVICE LEVEL OF admin;

-- VECTOR type and ANN search (ScyllaDB 5.4+)
-- CREATE TABLE items (id int PRIMARY KEY, embedding vector<float, 128>);
-- SELECT * FROM items ORDER BY embedding ANN OF [0.1, 0.2, ...] LIMIT 10;

-- DURATION type
-- CREATE TABLE events (id int PRIMARY KEY, duration duration);
-- INSERT INTO events (id, duration) VALUES (1, 1h30m);

-- GROUP BY (CQL extension)
SELECT id, COUNT(*) FROM events GROUP BY id;

-- LIKE operator
SELECT * FROM users WHERE name LIKE 'A%';
SELECT * FROM users WHERE name LIKE '%test%';

-- CAST
SELECT CAST(id AS text) FROM users;
SELECT CAST(created_at AS date) FROM users;

-- Static columns
CREATE TABLE sensor_data (sensor_id int, ts timestamp, location text STATIC, value double, PRIMARY KEY (sensor_id, ts));

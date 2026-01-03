-- Basic SELECT queries

SELECT * FROM users;
SELECT id, name FROM users;
SELECT id, name FROM keyspace1.users;
SELECT DISTINCT id FROM users;
SELECT COUNT(*) FROM users;
SELECT * FROM users WHERE id = 1;
SELECT * FROM users WHERE id = 1 AND name = 'test';
SELECT * FROM users WHERE id IN (1, 2, 3);
SELECT * FROM users WHERE id = 1 ORDER BY created_at DESC;
SELECT * FROM users WHERE id = 1 LIMIT 10;
-- NOTE: OFFSET is not part of CQL, use token-based pagination instead
SELECT * FROM users WHERE id = 1 ALLOW FILTERING;
SELECT JSON * FROM users;
SELECT toJson(name) FROM users;
SELECT writetime(name) FROM users;
SELECT ttl(name) FROM users;
SELECT * FROM users WHERE token(id) > token(1);

-- Function calls as values in WHERE clause
SELECT * FROM users WHERE created_at = currentTimestamp();
SELECT * FROM users WHERE created_at >= currentTimestamp();
SELECT * FROM users WHERE created_at < toTimestamp(now());
SELECT * FROM users WHERE id = uuid();
SELECT * FROM users WHERE event_time = now();
SELECT * FROM users WHERE created_at = currentDate();
SELECT * FROM users WHERE created_at = currentTime();
SELECT * FROM users WHERE id = 123 AND updated_at = currentTimestamp();
SELECT * FROM users WHERE token(id) >= token(uuid());

-- Aggregations
SELECT COUNT(*), AVG(age), SUM(balance) FROM users;
SELECT MIN(created_at), MAX(created_at) FROM users;

-- Aliases
SELECT id AS user_id, name AS user_name FROM users;

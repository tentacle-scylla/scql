-- Basic UPDATE queries

UPDATE users SET name = 'Bob' WHERE id = 1;
UPDATE users SET name = 'Bob', email = 'bob@example.com' WHERE id = 1;
UPDATE keyspace1.users SET name = 'Bob' WHERE id = 1;
UPDATE users SET name = 'Bob' WHERE id = 1 IF EXISTS;
UPDATE users SET name = 'Bob' WHERE id = 1 IF name = 'Alice';
UPDATE users USING TTL 3600 SET name = 'Bob' WHERE id = 1;
UPDATE users USING TIMESTAMP 1234567890 SET name = 'Bob' WHERE id = 1;

-- Counter updates
UPDATE counters SET count = count + 1 WHERE id = 1;
UPDATE counters SET count = count - 5 WHERE id = 1;

-- Collection updates
UPDATE users SET tags = tags + {'new_tag'} WHERE id = 1;
UPDATE users SET tags = tags - {'old_tag'} WHERE id = 1;
UPDATE users SET tags = {'a', 'b'} WHERE id = 1;
UPDATE users SET data['key'] = 'new_value' WHERE id = 1;
UPDATE users SET scores = scores + [4, 5] WHERE id = 1;
UPDATE users SET scores[0] = 10 WHERE id = 1;

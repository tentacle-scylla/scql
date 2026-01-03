-- Basic DELETE queries

DELETE FROM users WHERE id = 1;
DELETE FROM keyspace1.users WHERE id = 1;
DELETE name FROM users WHERE id = 1;
DELETE name, email FROM users WHERE id = 1;
DELETE FROM users WHERE id = 1 IF EXISTS;
DELETE FROM users WHERE id = 1 IF name = 'Alice';
DELETE FROM users USING TIMESTAMP 1234567890 WHERE id = 1;

-- Delete from collections
DELETE tags['old'] FROM users WHERE id = 1;
DELETE data['key'] FROM users WHERE id = 1;

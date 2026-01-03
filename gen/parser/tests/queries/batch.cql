-- Batch operations

BEGIN BATCH
  INSERT INTO users (id, name) VALUES (1, 'Alice');
  INSERT INTO users (id, name) VALUES (2, 'Bob');
APPLY BATCH;

BEGIN UNLOGGED BATCH
  INSERT INTO users (id, name) VALUES (1, 'Alice');
  UPDATE users SET name = 'Charlie' WHERE id = 2;
APPLY BATCH;

BEGIN BATCH USING TIMESTAMP 1234567890
  INSERT INTO users (id, name) VALUES (1, 'Alice');
  DELETE FROM users WHERE id = 3;
APPLY BATCH;

-- NOTE: COUNTER BATCH syntax not yet supported
-- BEGIN COUNTER BATCH
--   UPDATE counters SET count = count + 1 WHERE id = 1;
--   UPDATE counters SET count = count + 2 WHERE id = 2;
-- APPLY BATCH;

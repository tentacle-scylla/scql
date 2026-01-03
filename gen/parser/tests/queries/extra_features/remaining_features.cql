-- Remaining ScyllaDB features for 100% coverage

-- Service Level with SHARES option
CREATE SERVICE LEVEL gold WITH shares = 1000;
CREATE SERVICE LEVEL silver WITH timeout = 500ms AND shares = 500;
ALTER SERVICE LEVEL gold WITH shares = 2000;

-- HASHED PASSWORD in roles (for pre-hashed passwords)
CREATE ROLE service_account WITH HASHED PASSWORD = 'hashedvalue123';
CREATE ROLE batch_user WITH HASHED PASSWORD = 'abc123' AND LOGIN = true;
ALTER ROLE service_account WITH HASHED PASSWORD = 'newhash456';

-- EMPTY collection literal
INSERT INTO my_table (id, tags) VALUES (1, EMPTY);
UPDATE my_table SET tags = EMPTY WHERE id = 1;

-- PRUNE MATERIALIZED VIEW with CONCURRENCY
PRUNE MATERIALIZED VIEW my_view WHERE id = 1;
PRUNE MATERIALIZED VIEW ks.my_view WHERE id = 1 USING TIMEOUT 5000ms;
PRUNE MATERIALIZED VIEW my_view WHERE id = 1 USING CONCURRENCY 4;
PRUNE MATERIALIZED VIEW my_view WHERE id = 1 USING CONCURRENCY 4 AND TIMEOUT 5000ms;

-- VECTOR_SEARCH_INDEXING permission
GRANT VECTOR_SEARCH_INDEXING ON KEYSPACE my_keyspace TO my_role;
REVOKE VECTOR_SEARCH_INDEXING ON KEYSPACE my_keyspace FROM my_role;
LIST VECTOR_SEARCH_INDEXING ON KEYSPACE my_keyspace;

-- Internal debug functions (already partially tested, but ensure coverage)
SELECT SCYLLA_COUNTER_SHARD_LIST(counter_col) FROM counter_table WHERE id = 1;
SELECT SCYLLA_TIMEUUID_LIST_INDEX(timeuuid_col) FROM my_table WHERE id = 1;

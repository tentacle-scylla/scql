-- Internal/Debug ScyllaDB functions
-- These are typically used for debugging and internal operations

-- SCYLLA_COUNTER_SHARD_LIST - get counter shard details
SELECT SCYLLA_COUNTER_SHARD_LIST(counter_col) FROM counter_table WHERE id = 1;

-- SCYLLA_TIMEUUID_LIST_INDEX - get timeuuid list index
SELECT SCYLLA_TIMEUUID_LIST_INDEX(timeuuid_col) FROM my_table WHERE id = 1;

-- These functions are internal but may appear in queries
SELECT pk, SCYLLA_COUNTER_SHARD_LIST(value) FROM counters;

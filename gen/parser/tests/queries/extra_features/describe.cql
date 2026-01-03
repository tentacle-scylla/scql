-- DESCRIBE statements (server-side CQL since Cassandra 4.0 / ScyllaDB)
-- https://docs.scylladb.com/stable/cql/describe.html

-- Describe cluster
DESCRIBE CLUSTER;
DESC CLUSTER;

-- Describe schema
DESCRIBE SCHEMA;
DESC SCHEMA;
DESCRIBE FULL SCHEMA;
DESC FULL SCHEMA;

-- Describe keyspaces
DESCRIBE KEYSPACES;
DESC KEYSPACES;
DESCRIBE KEYSPACE;
DESC KEYSPACE;
DESCRIBE KEYSPACE my_keyspace;
DESC KEYSPACE my_keyspace;
DESCRIBE ONLY KEYSPACE my_keyspace;

-- Describe tables
DESCRIBE TABLES;
DESC TABLES;
DESCRIBE TABLE my_table;
DESC TABLE my_table;
DESCRIBE COLUMNFAMILY my_table;
DESC COLUMNFAMILY my_table;
DESCRIBE my_keyspace.my_table;

-- Describe indexes
DESCRIBE INDEX my_index;
DESC INDEX my_keyspace.my_index;

-- Describe materialized views
DESCRIBE MATERIALIZED VIEW my_view;
DESC MATERIALIZED VIEW my_keyspace.my_view;

-- Describe types
DESCRIBE TYPES;
DESC TYPES;
DESCRIBE TYPE my_type;
DESC TYPE my_keyspace.my_type;

-- Describe functions
DESCRIBE FUNCTIONS;
DESC FUNCTIONS;
DESCRIBE FUNCTION my_function;
DESC FUNCTION my_keyspace.my_function;

-- Describe aggregates
DESCRIBE AGGREGATES;
DESC AGGREGATES;
DESCRIBE AGGREGATE my_aggregate;
DESC AGGREGATE my_keyspace.my_aggregate;

-- Describe with internals (ScyllaDB extension for debugging)
DESCRIBE KEYSPACE my_keyspace WITH INTERNALS;
DESCRIBE TABLE my_table WITH INTERNALS;
DESCRIBE TABLE my_table WITH INTERNALS AND PASSWORDS;

-- Generic describe (auto-detect object type)
DESCRIBE my_object;
DESC my_keyspace.my_object;

-- Miscellaneous queries

-- USE statement
USE keyspace1;

-- NOTE: DESCRIBE/DESC are cqlsh client commands, not CQL statements
-- They are not parsed by the CQL parser

-- Data types in CREATE TABLE
CREATE TABLE all_types (
    id uuid PRIMARY KEY,
    col_ascii ascii,
    col_bigint bigint,
    col_blob blob,
    col_boolean boolean,
    col_counter counter,
    col_date date,
    col_decimal decimal,
    col_double double,
    col_float float,
    col_inet inet,
    col_int int,
    col_smallint smallint,
    col_text text,
    col_time time,
    col_timestamp timestamp,
    col_timeuuid timeuuid,
    col_tinyint tinyint,
    col_uuid uuid,
    col_varchar varchar,
    col_varint varint,
    col_frozen frozen<address>,
    col_list list<text>,
    col_set set<int>,
    col_map map<text, int>,
    col_tuple tuple<int, text, double>
);

-- Prepared statement markers
SELECT * FROM users WHERE id = ?;
INSERT INTO users (id, name) VALUES (?, ?);
UPDATE users SET name = ? WHERE id = ?;
DELETE FROM users WHERE id = ?;

-- Named markers
SELECT * FROM users WHERE id = :id;
INSERT INTO users (id, name) VALUES (:id, :name);

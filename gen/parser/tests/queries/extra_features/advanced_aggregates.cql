-- Advanced User-Defined Aggregates with REDUCEFUNC
-- https://docs.scylladb.com/stable/cql/functions.html#user-defined-aggregates

-- Create state function for aggregate
CREATE FUNCTION state_sum(state int, val int)
    CALLED ON NULL INPUT
    RETURNS int
    LANGUAGE lua
    AS 'return state + val';

-- Create reduce function (for distributed aggregation)
CREATE FUNCTION reduce_sum(state1 int, state2 int)
    CALLED ON NULL INPUT
    RETURNS int
    LANGUAGE lua
    AS 'return state1 + state2';

-- Create final function
CREATE FUNCTION finalize_avg(state tuple<int, int>)
    CALLED ON NULL INPUT
    RETURNS double
    LANGUAGE lua
    AS 'return state[0] / state[1]';

-- Create aggregate with REDUCEFUNC
CREATE AGGREGATE my_sum(int)
    SFUNC state_sum
    STYPE int
    REDUCEFUNC reduce_sum
    INITCOND 0;

-- Create aggregate with all options
CREATE AGGREGATE my_avg(int)
    SFUNC state_avg_add
    STYPE tuple<int, int>
    REDUCEFUNC state_avg_reduce
    FINALFUNC finalize_avg
    INITCOND (0, 0);

-- Create aggregate if not exists
CREATE AGGREGATE IF NOT EXISTS custom_count(int)
    SFUNC count_state
    STYPE int
    INITCOND 0;

-- Replace aggregate
CREATE OR REPLACE AGGREGATE my_sum(int)
    SFUNC state_sum
    STYPE int
    INITCOND 0;

-- Drop aggregate
DROP AGGREGATE my_sum;
DROP AGGREGATE IF EXISTS my_avg;
DROP AGGREGATE my_keyspace.my_sum;
DROP AGGREGATE my_sum(int);

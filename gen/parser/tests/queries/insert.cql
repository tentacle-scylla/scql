-- Basic INSERT queries

INSERT INTO users (id, name) VALUES (1, 'Alice');
INSERT INTO users (id, name, email) VALUES (1, 'Alice', 'alice@example.com');
INSERT INTO keyspace1.users (id, name) VALUES (1, 'Alice');
INSERT INTO users (id, name) VALUES (1, 'Alice') IF NOT EXISTS;
INSERT INTO users (id, name) VALUES (1, 'Alice') USING TTL 3600;
INSERT INTO users (id, name) VALUES (1, 'Alice') USING TIMESTAMP 1234567890;
INSERT INTO users (id, name) VALUES (1, 'Alice') USING TTL 3600 AND TIMESTAMP 1234567890;

-- JSON insert
INSERT INTO users JSON '{"id": 1, "name": "Alice"}';
-- NOTE: DEFAULT UNSET/NULL syntax not yet supported
-- INSERT INTO users JSON '{"id": 1, "name": "Alice"}' DEFAULT UNSET;

-- With collections
INSERT INTO users (id, tags) VALUES (1, {'a', 'b', 'c'});
INSERT INTO users (id, data) VALUES (1, {'key': 'value'});
INSERT INTO users (id, scores) VALUES (1, [1, 2, 3]);

-- NOTE: UDT literal syntax {field: value} not yet supported
-- INSERT INTO users (id, address) VALUES (1, {street: '123 Main St', city: 'NYC'});

-- With tuple
INSERT INTO users (id, location) VALUES (1, (40.7128, -74.0060));

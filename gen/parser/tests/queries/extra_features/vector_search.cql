-- Vector Search / ANN (Approximate Nearest Neighbor)
-- https://docs.scylladb.com/stable/cql/vector-search.html

-- Create table with vector column
CREATE TABLE embeddings (
    id uuid PRIMARY KEY,
    embedding vector<float, 128>,
    metadata text
);

CREATE TABLE items (
    category text,
    item_id uuid,
    name text,
    description_embedding vector<float, 768>,
    PRIMARY KEY (category, item_id)
);

-- Create vector index (SAI - Storage Attached Index)
CREATE CUSTOM INDEX ON embeddings(embedding) USING 'StorageAttachedIndex';
CREATE CUSTOM INDEX embedding_idx ON embeddings(embedding) USING 'StorageAttachedIndex';
CREATE CUSTOM INDEX ON embeddings(embedding) USING 'StorageAttachedIndex' WITH OPTIONS = {'similarity_function': 'cosine'};
CREATE CUSTOM INDEX ON embeddings(embedding) USING 'StorageAttachedIndex' WITH OPTIONS = {'similarity_function': 'euclidean'};
CREATE CUSTOM INDEX ON embeddings(embedding) USING 'StorageAttachedIndex' WITH OPTIONS = {'similarity_function': 'dot_product'};

-- ANN queries - ORDER BY ... ANN OF ...
SELECT * FROM embeddings ORDER BY embedding ANN OF [0.1, 0.2, 0.3] LIMIT 10;
SELECT id, metadata FROM embeddings ORDER BY embedding ANN OF [0.5, 0.5, 0.5, 0.5] LIMIT 5;
SELECT * FROM items WHERE category = 'electronics' ORDER BY description_embedding ANN OF [0.1, 0.2, 0.3] LIMIT 10;

-- ANN with filtering
SELECT * FROM embeddings WHERE id > 123e4567-e89b-12d3-a456-426614174000 ORDER BY embedding ANN OF [0.1, 0.2] LIMIT 10;

-- Insert vector data
INSERT INTO embeddings (id, embedding, metadata) VALUES (uuid(), [0.1, 0.2, 0.3, 0.4], 'test document');
INSERT INTO embeddings (id, embedding) VALUES (now(), [1.0, 2.0, 3.0, 4.0, 5.0]);

-- Update vector data
UPDATE embeddings SET embedding = [0.5, 0.6, 0.7, 0.8] WHERE id = 123e4567-e89b-12d3-a456-426614174000;

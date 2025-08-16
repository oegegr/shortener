-- migrations/000001_create_url_table.up.sql
CREATE TABLE url (
    id SERIAL PRIMARY KEY,
    short_id VARCHAR(255) NOT NULL,
    url VARCHAR(255) NOT NULL
);

CREATE INDEX idx_short_id ON url(short_id);
-- +goose Up
CREATE TABLE departments (
    id SERIAL PRIMARY KEY,
    name VARCHAR(200) NOT NULL,
    parent_id INTEGER REFERENCES departments(id) ON DELETE RESTRICT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_departments_parent_name ON departments (parent_id, name);
CREATE INDEX idx_departments_parent_id ON departments (parent_id);

-- +goose Down
DROP TABLE departments;
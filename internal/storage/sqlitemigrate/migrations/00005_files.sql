-- +goose Up
CREATE TABLE vault_files (
    id TEXT PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    user_id TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    version BIGINT NOT NULL DEFAULT 0,
    metadata TEXT NOT NULL DEFAULT '{}',
    salt VARCHAR(255),
    iv VARCHAR(255),
    filename VARCHAR(255),
    location VARCHAR(255),
    checksum VARCHAR(255),
    size BIGINT
);

CREATE UNIQUE INDEX vault_files_user_id_name_idx ON vault_files (user_id, name);


-- +goose Down
DROP INDEX vault_files_user_id_name_idx;
DROP TABLE vault_files;

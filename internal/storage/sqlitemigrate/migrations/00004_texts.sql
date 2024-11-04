-- +goose Up
CREATE TABLE vault_texts (
    id TEXT PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    user_id TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    version BIGINT NOT NULL DEFAULT 0,
    metadata TEXT NOT NULL DEFAULT '{}',
    salt VARCHAR(255),
    iv VARCHAR(255),
    location VARCHAR(255),
    checksum VARCHAR(255)
);

CREATE UNIQUE INDEX vault_texts_user_id_name_idx ON vault_texts (user_id, name);


-- +goose Down
DROP INDEX vault_texts_user_id_name_idx;
DROP TABLE vault_texts;

-- +goose Up
CREATE TABLE vault_credentials (
    id TEXT PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    user_id TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    version BIGINT NOT NULL DEFAULT 0,
    metadata TEXT NOT NULL DEFAULT '{}',
    data TEXT NOT NULL DEFAULT '{}'
);

CREATE UNIQUE INDEX vault_credentials_user_id_name_idx ON vault_credentials (user_id, name);


-- +goose Down
DROP INDEX vault_credentials_user_id_name_idx;
DROP TABLE vault_credentials;

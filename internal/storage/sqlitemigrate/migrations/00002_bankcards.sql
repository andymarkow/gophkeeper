-- +goose Up
CREATE TABLE vault_bankcards (
    id TEXT PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    user_id UUID NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    version BIGINT NOT NULL DEFAULT 0,
    metadata TEXT NOT NULL DEFAULT '{}',
    data TEXT NOT NULL DEFAULT '{}'
);

CREATE UNIQUE INDEX vault_bankcards_user_id_name_idx ON vault_bankcards (user_id, name);


-- +goose Down
DROP INDEX vault_bankcards_user_id_name_idx;
DROP TABLE vault_bankcards;

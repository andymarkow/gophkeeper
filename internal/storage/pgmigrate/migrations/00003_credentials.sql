-- +goose Up
CREATE TABLE vault_credentials (
    id UUID PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    user_id UUID NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    data JSONB NOT NULL DEFAULT '{}'::jsonb
);

CREATE UNIQUE INDEX vault_credentials_user_id_name_idx ON vault_credentials (user_id, name);


-- +goose Down
DROP INDEX vault_credentials_user_id_name_idx;
DROP TABLE vault_credentials;

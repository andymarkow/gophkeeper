-- +goose Up
CREATE TABLE vault_texts (
    id UUID PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    user_id UUID NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    salt VARCHAR(255),
    iv VARCHAR(255),
    location VARCHAR(255),
    checksum VARCHAR(255)
);

CREATE UNIQUE INDEX vault_texts_user_id_name_idx ON vault_texts (user_id, name);


-- +goose Down
DROP INDEX vault_texts_user_id_name_idx;
DROP TABLE vault_texts;

-- +goose Up
CREATE TABLE users (
            id UUID PRIMARY KEY,
            login VARCHAR(255) NOT NULL UNIQUE,
            password VARCHAR(255) NOT NULL
        );

CREATE INDEX IF NOT EXISTS users_login_idx ON users (login);


-- +goose Down
DROP INDEX users_login_idx;
DROP TABLE users;

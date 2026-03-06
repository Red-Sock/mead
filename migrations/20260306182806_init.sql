-- +goose Up
-- +goose StatementBegin
CREATE TABLE users
(
    username TEXT PRIMARY KEY,
    pass     TEXT NOT NULL 
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE users;
-- +goose StatementEnd

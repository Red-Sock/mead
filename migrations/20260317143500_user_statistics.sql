-- +goose Up
-- +goose StatementBegin
CREATE TABLE user_statistics (
    username TEXT PRIMARY KEY,
    last_connect TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    bytes_passed BIGINT DEFAULT 0,
    FOREIGN KEY(username) REFERENCES users(username) ON DELETE CASCADE
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE user_statistics;
-- +goose StatementEnd

-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS extra_proxies
(
    proxy_link TEXT UNIQUE NOT NULL 
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS extra_proxies;
-- +goose StatementEnd

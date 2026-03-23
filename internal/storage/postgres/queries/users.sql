-- name: Add :exec
INSERT INTO users (username, pass, telegram_id)
VALUES ($1, $2, $3);

-- name: GetByTelegramName :one
SELECT username, telegram_id
FROM users
WHERE username = $1
LIMIT 1;

-- name: GetByTelegramId :one
SELECT username, telegram_id
FROM users
WHERE telegram_id = $1
LIMIT 1;

-- name: GetPassByUsername :one
SELECT pass
FROM users
WHERE username = $1
LIMIT 1;

-- name: UpdateStatistics :exec
INSERT INTO user_statistics (username, last_connect, bytes_passed)
VALUES ($1, CURRENT_TIMESTAMP, $2)
ON CONFLICT(username) DO UPDATE SET
    last_connect = excluded.last_connect,
    bytes_passed = user_statistics.bytes_passed + excluded.bytes_passed;

-- name: ListStatistics :many
SELECT username, last_connect, bytes_passed
FROM user_statistics;
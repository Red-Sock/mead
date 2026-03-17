-- name: Add :exec
INSERT INTO users (username, pass, telegram_id)
VALUES (?, ?, ?);

-- name: GetByTelegramName :one
SELECT username, telegram_id
FROM users
WHERE username = ?
LIMIT 1;

-- name: GetByTelegramId :one
SELECT username, telegram_id
FROM users
WHERE telegram_id = ?
LIMIT 1;

-- name: GetPassByUsername :one
SELECT pass
FROM users
WHERE username = ?
LIMIT 1;

-- name: UpdateStatistics :exec
INSERT INTO user_statistics (username, last_connect, bytes_passed)
VALUES (?, CURRENT_TIMESTAMP, ?)
ON CONFLICT(username) DO UPDATE SET
    last_connect = excluded.last_connect,
    bytes_passed = user_statistics.bytes_passed + excluded.bytes_passed;
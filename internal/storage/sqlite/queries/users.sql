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
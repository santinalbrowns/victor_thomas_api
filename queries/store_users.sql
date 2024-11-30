-- name: AssignStoreUser :exec
INSERT INTO store_users (store_id, user_id) VALUES (?, ?);

-- name: CheckStoreUser :one
SELECT COUNT(*) > 0
FROM store_users
WHERE store_id = ? AND user_id = ?;

-- name: FindStoreUsers :many
SELECT u.*
FROM store_users AS su
JOIN users AS u ON u.id = su.user_id
WHERE su.store_id = ?;

-- name: DeleteStoreUser :exec
DELETE FROM store_users
WHERE store_id = ? AND user_id = ?;
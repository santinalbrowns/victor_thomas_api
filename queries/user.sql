-- name: InsertUser :execlastid
INSERT INTO users (firstname, lastname, email, password) VALUES (?, ?, ?, ?);

-- name: FindUserByEmail :one
SELECT * FROM users WHERE email = ?;

-- name: FindUserByID :one
SELECT * FROM users WHERE id = ? LIMIT 1;

-- name: FindUserStore :one
SELECT s.* FROM stores s
JOIN store_users su ON su.store_id = s.id
JOIN users u ON u.id = su.user_id 
WHERE u.id = ?
LIMIT 1;

-- name: FindUserStores :many
SELECT s.* FROM stores s
JOIN store_users su ON su.store_id = s.id
JOIN users u ON u.id = su.user_id 
WHERE u.id = ?;
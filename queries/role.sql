-- name: InsertRole :execlastid
INSERT INTO roles (name) VALUES (?);

-- name: FindRole :one
SELECT * FROM roles WHERE id = ?;

-- name: FindRoleByName :one
SELECT * FROM roles WHERE name = ?;
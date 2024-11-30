-- name: AssignUserRole :exec
INSERT INTO user_roles (user_id, role_id) VALUES (?, ?);

-- name: CheckUserRole :one
SELECT COUNT(*) > 0
FROM user_roles
WHERE user_id = ? AND role_id = ?;

-- name: FindUserRoles :many
SELECT r.id, r.name
FROM roles AS r
JOIN user_roles AS ur ON ur.role_id = r.id
WHERE ur.user_id = ?;
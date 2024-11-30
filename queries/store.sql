-- name: InsertStore :execlastid
INSERT INTO stores (slug, name, status)
VALUES (?, ?, ?);

-- name: FindStore :one
SELECT * FROM stores
WHERE id = ?;

-- name: FindStoreBySlug :one
SELECT * FROM stores
WHERE slug = ?;

-- name: FindStores :many
SELECT * FROM stores
ORDER BY id DESC
LIMIT ? OFFSET ?;

-- name: SearchStores :many
SELECT * FROM stores
WHERE name LIKE ?
ORDER BY id DESC
LIMIT ? OFFSET ?;

-- name: CountStores :one
SELECT COUNT(*) AS count
FROM stores;

-- name: UpdateStore :exec
UPDATE stores
SET 
    name = COALESCE(?, name),
    slug = COALESCE(?, slug),
    status = COALESCE(?, status)
WHERE id = ?;

-- name: DeleteStore :exec
DELETE FROM stores
WHERE id = ?;


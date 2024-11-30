-- name: InsertImage :execlastid
INSERT INTO images (name) VALUES (?);

-- name: FindImage :one
SELECT * FROM images WHERE id = ?;

-- name: DeleteImage :exec
DELETE FROM images WHERE id = ?;
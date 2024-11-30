-- name: AssignProductImage :exec
INSERT INTO product_images (product_id, image_id) VALUES (?, ?);

-- name: FindProductImages :many
SELECT i.id, i.name
FROM images AS i
JOIN product_images AS pi ON pi.image_id = i.id
WHERE pi.product_id = ?;
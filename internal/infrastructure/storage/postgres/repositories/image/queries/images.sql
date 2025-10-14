-- name: GetImageByID :one
SELECT * FROM images
WHERE id = $1 LIMIT 1;

-- name: GetImageWithProcessedData :one
SELECT
    i.*,
    p.width,
    p.height,
    p.processed_at
FROM images i
         LEFT JOIN processed_images p ON i.id = p.image_id
WHERE i.id = $1;

-- name: CreateImage :one
INSERT INTO images (
    id, original_name, file_name, status, result_url, size, format, uploaded_at
) VALUES (
             $1, $2, $3, $4, $5, $6, $7, $8
         )
    RETURNING *;

-- name: UpdateImageStatus :one
UPDATE images
SET status = $2
WHERE id = $1
    RETURNING *;

-- name: UpdateImage :one
UPDATE images
SET
    status = COALESCE($2, status),
    result_url = COALESCE($3, result_url)
WHERE id = $1
    RETURNING *;

-- name: DeleteImage :exec
DELETE FROM images
WHERE id = $1;

-- name: ListImages :many
SELECT * FROM images
ORDER BY uploaded_at DESC
    LIMIT $1 OFFSET $2;

-- name: ListImagesByStatus :many
SELECT * FROM images
WHERE status = $1
ORDER BY uploaded_at DESC
    LIMIT $2 OFFSET $3;

-- name: ListImagesWithFilters :many
SELECT * FROM images
WHERE
    (sqlc.narg('status')::VARCHAR IS NULL OR status = sqlc.narg('status')) AND
    (sqlc.narg('format')::VARCHAR IS NULL OR format = sqlc.narg('format')) AND
    (sqlc.narg('from_date')::TIMESTAMP IS NULL OR uploaded_at >= sqlc.narg('from_date')) AND
    (sqlc.narg('to_date')::TIMESTAMP IS NULL OR uploaded_at <= sqlc.narg('to_date'))
ORDER BY uploaded_at DESC
    LIMIT sqlc.narg('limit') OFFSET sqlc.narg('offset');

-- name: CountImagesWithFilters :one
SELECT COUNT(*) FROM images
WHERE
    (sqlc.narg('status')::VARCHAR IS NULL OR status = sqlc.narg('status')) AND
    (sqlc.narg('format')::VARCHAR IS NULL OR format = sqlc.narg('format')) AND
    (sqlc.narg('from_date')::TIMESTAMP IS NULL OR uploaded_at >= sqlc.narg('from_date')) AND
    (sqlc.narg('to_date')::TIMESTAMP IS NULL OR uploaded_at <= sqlc.narg('to_date'));

-- name: GetImagesByFileName :many
SELECT * FROM images
WHERE file_name = $1
ORDER BY uploaded_at DESC;

-- name: CreateProcessedImage :one
INSERT INTO processed_images (
    image_id, width, height, processed_at
) VALUES (
             $1, $2, $3, $4
         )
    RETURNING *;

-- name: GetProcessedImage :one
SELECT * FROM processed_images
WHERE image_id = $1 LIMIT 1;

-- name: UpdateProcessedImage :one
UPDATE processed_images
SET
    width = $2,
    height = $3,
    processed_at = $4
WHERE image_id = $1
    RETURNING *;

-- name: DeleteProcessedImage :exec
DELETE FROM processed_images
WHERE image_id = $1;

-- name: GetRecentProcessedImages :many
SELECT
    i.*,
    p.width,
    p.height,
    p.processed_at
FROM images i
         JOIN processed_images p ON i.id = p.image_id
WHERE p.processed_at >= $1
ORDER BY p.processed_at DESC
    LIMIT $2;

-- name: GetImagesStats :one
SELECT
    COUNT(*) as total_images,
    COUNT(CASE WHEN status = 'completed' THEN 1 END) as completed_count,
    COUNT(CASE WHEN status = 'processing' THEN 1 END) as processing_count,
    COUNT(CASE WHEN status = 'failed' THEN 1 END) as failed_count,
    COALESCE(SUM(size), 0) as total_size
FROM images;
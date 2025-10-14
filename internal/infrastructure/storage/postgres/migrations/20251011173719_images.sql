-- +goose Up
-- +goose StatementBegin
CREATE TABLE images (
    id UUID PRIMARY KEY,
    original_name VARCHAR NOT NULL,
    file_name VARCHAR NOT NULL,
    status VARCHAR NOT NULL, -- "uploaded", "processing", "completed", "failed"
    result_url VARCHAR,
    size BIGINT NOT NULL,
    format VARCHAR NOT NULL,
    uploaded_at TIMESTAMP NOT NULL
);

CREATE TABLE processed_images (
    image_id UUID PRIMARY KEY REFERENCES images(id) ON DELETE CASCADE,
    width INT NOT NULL,
    height INT NOT NULL,
    processed_at TIMESTAMP NOT NULL
);

CREATE INDEX idx_images_status ON images(status);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS processed_images CASCADE;
DROP TABLE IF EXISTS images CASCADE;
-- +goose StatementEnd
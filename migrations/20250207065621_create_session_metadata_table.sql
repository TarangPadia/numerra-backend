-- +goose Up
CREATE TABLE session_metadata (
    session_id INT AUTO_INCREMENT PRIMARY KEY,
    user_email VARCHAR(255) NOT NULL UNIQUE,
    selected_org_id CHAR(36) NULL
);

-- +goose Down
DROP TABLE session_metadata;

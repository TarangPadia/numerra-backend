-- +goose Up
CREATE TABLE users (
    id CHAR(36) PRIMARY KEY,
    email VARCHAR(32) UNIQUE NOT NULL,
    first_name VARCHAR(32) NOT NULL,
    last_name VARCHAR(32) NOT NULL,
    show_welcome_prompt BOOLEAN DEFAULT TRUE,
    is_email_verified BOOLEAN DEFAULT FALSE
);

-- +goose Down
DROP TABLE users;
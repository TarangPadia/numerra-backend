-- +goose Up
CREATE TABLE password_reset_codes (
  id CHAR(36) PRIMARY KEY,
  email VARCHAR(255) NOT NULL,
  code TEXT NOT NULL,
  expires_at DATETIME NOT NULL
);

-- +goose Down
DROP TABLE password_reset_codes;

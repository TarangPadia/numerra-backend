-- +goose Up
CREATE TABLE user_otps (
    id INT AUTO_INCREMENT PRIMARY KEY,
    email VARCHAR(255) NOT NULL,
    otp VARCHAR(10) NOT NULL,
    expires_at DATETIME NOT NULL
);

-- +goose Down
DROP TABLE user_otps;

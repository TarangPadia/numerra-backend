-- +goose Up
CREATE TABLE invitations (
    id CHAR(36) PRIMARY KEY,
    user_email VARCHAR(255) NOT NULL,
    organization_id CHAR(36) NOT NULL,
    inviter_user_id CHAR(36) NOT NULL,
    is_accepted BOOLEAN DEFAULT FALSE,
    expires_at DATETIME NOT NULL
);

-- +goose Down
DROP TABLE invitations;

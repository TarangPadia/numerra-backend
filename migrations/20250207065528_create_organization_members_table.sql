-- +goose Up
CREATE TABLE organization_members (
    member_id CHAR(36) PRIMARY KEY,
    user_id CHAR(36) NOT NULL,
    organization_id CHAR(36) NOT NULL,
    role ENUM('ROLE_OWNER','ROLE_ADMIN','ROLE_EDITOR','ROLE_SPECTATOR') NOT NULL DEFAULT 'ROLE_SPECTATOR',
    joined_at DATETIME NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (organization_id) REFERENCES organizations(organization_id) ON DELETE CASCADE
);

-- +goose Down
DROP TABLE organization_members;

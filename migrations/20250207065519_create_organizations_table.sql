-- +goose Up
CREATE TABLE organizations (
    organization_id CHAR(36) PRIMARY KEY,
    organization_name VARCHAR(255) NOT NULL,
    incorporation_state VARCHAR(255),
    incorporation_year INT,
    industry VARCHAR(255),
    revenue DECIMAL(15,2)
);

-- +goose Down
DROP TABLE organizations;

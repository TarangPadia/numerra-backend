-- +goose Up
CREATE TABLE organization_integrations (
  id CHAR(36) PRIMARY KEY,
  organization_id CHAR(36) NOT NULL,
  provider VARCHAR(32) NOT NULL,

  status ENUM('CONNECTED','DISCONNECTED','NEEDS_REAUTH') NOT NULL DEFAULT 'DISCONNECTED',

  external_account_id VARCHAR(255) NULL,
  scopes TEXT NULL,

  access_token_enc TEXT NULL,
  refresh_token_enc TEXT NULL,

  access_expires_at DATETIME NULL,
  refresh_expires_at DATETIME NULL,

  connected_by_user_id CHAR(36) NULL,
  connected_at DATETIME NULL,
  last_refreshed_at DATETIME NULL,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

  UNIQUE KEY uq_org_provider (organization_id, provider),
  INDEX idx_org_provider_status (organization_id, provider, status),

  FOREIGN KEY (organization_id) REFERENCES organizations(organization_id) ON DELETE CASCADE
);

-- +goose Down
DROP TABLE organization_integrations;

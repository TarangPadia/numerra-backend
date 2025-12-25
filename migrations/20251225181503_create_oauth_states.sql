-- +goose Up
CREATE TABLE oauth_states (
  id CHAR(36) PRIMARY KEY,
  organization_id CHAR(36) NOT NULL,
  provider VARCHAR(32) NOT NULL,
  state VARCHAR(128) NOT NULL,
  created_by_user_id CHAR(36) NOT NULL,
  expires_at DATETIME NOT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,

  INDEX idx_oauth_states_state_provider (state, provider),
  INDEX idx_oauth_states_org_provider (organization_id, provider),

  FOREIGN KEY (organization_id) REFERENCES organizations(organization_id) ON DELETE CASCADE
);

-- +goose Down
DROP TABLE oauth_states;


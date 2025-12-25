-- +goose Up
ALTER TABLE organization_members
  ADD CONSTRAINT fk_om_org
  FOREIGN KEY (organization_id) REFERENCES organizations(organization_id)
  ON DELETE CASCADE;

-- +goose Down
ALTER TABLE organization_members DROP FOREIGN KEY fk_om_org;

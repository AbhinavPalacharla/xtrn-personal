-- +goose Up
-- +goose StatementBegin
CREATE TABLE oauth_providers (
  name TEXT PRIMARY KEY,
  client_id TEXT NOT NULL,
  client_secret TEXT NOT NULL,
  callback_url TEXT NOT NULL,
  scopes TEXT
);

CREATE TABLE oauth_tokens (
  id TEXT PRIMARY KEY,
  refresh_token TEXT UNIQUE NOT NULL,
  oauth_provider TEXT NOT NULL,
  FOREIGN KEY (oauth_provider) REFERENCES oauth_providers (name)
);

CREATE TABLE mcp_server_images (
  id TEXT UNIQUE NOT NULL,
  slug TEXT NOT NULL,
  version INTEGER NOT NULL,
  name TEXT NOT NULL,
  docker_image TEXT NOT NULL,
  type TEXT NOT NULL CHECK (
    type in ('PUBLIC', 'AUTHENTICATED_OAUTH', 'AUTHENTICATED')
  ),
  oauth_provider TEXT,
  env_schema JSON NOT NULL,
  PRIMARY KEY (slug, version),
  FOREIGN KEY (oauth_provider) REFERENCES oauth_providers (name)
);

CREATE TABLE mcp_server_tools (
  id TEXT NOT NULL,
  name TEXT NOT NULL,
  description TEXT,
  schema TEXT NOT NULL,
  image_id TEXT NOT NULL,
  FOREIGN KEY (image_id) REFERENCES mcp_server_images (id)
);

CREATE TABLE mcp_server_instances (
  id TEXT PRIMARY KEY,
  slug TEXT NOT NULL,
  version INTEGER NOT NULL,
  address TEXT NOT NULL,
  env JSON NOT NULL,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (slug, version) REFERENCES mcp_server_images (slug, version)
);

-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS oauth_tokens;

DROP TABLE IF EXISTS oauth_providers;

DROP TABLE IF EXISTS mcp_server_instances;

DROP TABLE IF EXISTS mcp_server_images;

-- +goose StatementEnd

/***********************************/
/*
OAUTH token queries
*/
-- name: InsertOauthProvider :exec
INSERT INTO
  oauth_providers (name, client_id, client_secret)
VALUES
  (?, ?, ?);

-- name: InsertOauthToken :exec
INSERT INTO
  oauth_tokens (
    id,
    access_token,
    refresh_token,
    expiry,
    oauth_provider
  )
VALUES
  (?, ?, ?, ?, ?);

/***********************************/
/*
MCP Server Image Queries
*/
-- name: InsertMCPServerImage :exec
INSERT INTO
  mcp_server_images (
    id,
    slug,
    version,
    name,
    docker_image,
    type,
    oauth_provider,
    env_schema
  )
VALUES
  (?, ?, ?, ?, ?, ?, ?, ?);

-- name: GetMCPServerImage :one
SELECT
  images.*,
  providers.name as provider_name,
  providers.client_id,
  providers.client_secret
FROM
  mcp_server_images AS images
  LEFT JOIN oauth_providers AS providers ON images.oauth_provider = providers.name
WHERE
  images.id = ?;

/***********************************/
/*
MCP Server Instance Queries
*/
-- name: InsertMCPServerInstance :exec
INSERT INTO
  mcp_server_instances (id, slug, version, address, env)
VALUES
  (?, ?, ?, ?, ?);

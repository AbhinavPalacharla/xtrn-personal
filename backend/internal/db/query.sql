/***********************************/
/*
OAUTH token queries
*/
-- name: InsertOauthProvider :exec
INSERT INTO
  oauth_providers (
    name,
    client_id,
    client_secret,
    callback_url,
    scopes
  )
VALUES
  (?, ?, ?, ?, ?);

-- name: InsertOauthToken :exec
INSERT INTO
  oauth_tokens (id, refresh_token, oauth_provider)
VALUES
  (?, ?, ?);

-- name: GetOauthTokenByProvider :one
SELECT
  *
FROM
  oauth_tokens
WHERE
  oauth_tokens.oauth_provider = ?;

-- name: UpdateOauthTokenByProivder :exec
UPDATE oauth_tokens
SET
  refresh_token = ?
WHERE
  oauth_provider = ?;

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

-- name: InsertMCPServerInstanceTool :exec
INSERT INTO
  mcp_server_tools (id, name, description, schema, image_id)
VALUES
  (?, ?, ?, ?, ?);

-- name: DeleteMCPServerInstance :exec
DELETE FROM mcp_server_instances
WHERE
  id = ?;

-- name: GetMCPServerInstances :many
SELECT
  inst.id as instance_id,
  inst.address,
  img.id AS image_id,
  tool.name as tool_name,
  tool.description as tool_desc,
  tool.schema as tool_schema
FROM
  mcp_server_instances inst
  LEFT JOIN mcp_server_images AS img ON inst.slug = img.slug
  AND inst.version = img.version
  LEFT JOIN mcp_server_tools as tool ON img.id = tool.image_id;

-- name: DeleteAllMCPinstances :exec
DELETE FROM mcp_server_instances;

/***********************************/
/*
Chat Queries
*/
-- name: InsertChat :exec
INSERT INTO
  chats (id)
VALUES
  (?);

-- name: InsertMessage :exec
INSERT INTO
  messages (id, role, content, stop_reason, chat_id)
VALUES
  (?, ?, ?, ?, ?);

-- name: InsertToolCallRequest :exec
INSERT INTO
  tool_call_request (message_id, tool_call_id, name, arguments)
VALUES
  (?, ?, ?, ?);

-- name: InsertToolCallResult :exec
INSERT INTO
  tool_call_result (message_id, tool_call_id, name, content, is_error)
VALUES
  (?, ?, ?, ?, ?);

-- name: GetMessages :many
SELECT
  m.id,
  m.role,
  m.content,
  m.stop_reason,
  m.chat_id,
  treq.tool_call_id as treq_tool_call_id,
  treq.name as treq_name,
  treq.arguments as treq_args,
  tres.tool_call_id as tres_id,
  tres.name as tres_name,
  tres.content as tres_content,
  tres.is_error as tres_is_error
FROM
  messages m
  LEFT JOIN tool_call_request as treq ON treq.message_id = m.id
  LEFT JOIN tool_call_result as tres ON tres.message_id = m.id
WHERE
  chat_id = ?;

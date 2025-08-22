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

-- name: InsertAIMessagePart :one
INSERT INTO
  ai_message_parts (type, part_index, message_id)
VALUES
  (?, ?, ?) RETURNING id;

-- name: InsertTextPart :exec
INSERT INTO
  text_part (text, message_part_id)
VALUES
  (?, ?);

-- name: InsertToolCallPart :exec
INSERT INTO
  tool_call_part (tool_call_id, name, arguments, message_part_id)
VALUES
  (?, ?, ?, ?);

-- name: InsertToolCallResult :exec
INSERT INTO
  tool_call_result (message_id, tool_call_id, name, content, is_error)
VALUES
  (?, ?, ?, ?, ?);

-- name: InsertAuthRequest :exec
INSERT INTO
  chat_auth_requests (id, status, oauth_provider_name, chat_id)
VALUES
  (?, ?, ?, ?);

/***********************************/
-- name: GetViewChatMessges :many
SELECT
  *
FROM
  v_get_chat_messages
WHERE
  chat_id = ?;

-- name: GetChatWithAuthAndMessages :one
SELECT
  *
FROM
  v_get_chat_with_auth_and_messages
WHERE
  chat_id = ?;

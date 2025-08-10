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

/***********************************/
-- name: GetChatMessages :many
SELECT
  m.*,
  CASE
    WHEN m.role = 'ai' THEN (
      SELECT
        json_group_array(p.part_json)
      FROM
        (
          SELECT
            CASE
              WHEN amp.type = 'text' THEN json_object(
                'type',
                'text',
                'index',
                amp.part_index,
                'text',
                tp.text
              )
              WHEN amp.type = 'function' THEN json_object(
                'type',
                'tool_call',
                'index',
                amp.part_index,
                'tool_call_id',
                tcp.tool_call_id,
                'name',
                tcp.name,
                'arguments',
                tcp.arguments
              )
            END AS part_json
          FROM
            ai_message_parts amp
            LEFT JOIN text_part tp ON tp.message_part_id = amp.id
            AND amp.type = 'text'
            LEFT JOIN tool_call_part tcp ON tcp.message_part_id = amp.id
            AND amp.type = 'function'
          WHERE
            amp.message_id = m.id
          ORDER BY
            amp.part_index,
            amp.id
        ) p
    )
  END AS ai_message,
  CASE
    WHEN m.role = 'tool' THEN (
      SELECT
        json_object(
          'tool_call_id',
          t.tool_call_id,
          'name',
          t.name,
          'content',
          t.content,
          'is_error',
          t.is_error
        )
      FROM
        tool_call_result t
      WHERE
        t.message_id = m.id
    )
  END AS tool_result
FROM
  messages m
WHERE
  m.chat_id = ?
ORDER BY
  m.id;

-- name: GetChatsWithMessageCount :many
SELECT
  c.id,
  (
    SELECT
      COUNT(*)
    FROM
      messages m
    WHERE
      m.chat_id = c.id
  ) as message_count
FROM
  chats c
WHERE
  c.id IN (
    SELECT
      chat_id
    FROM
      messages
    GROUP BY
      chat_id
    HAVING
      COUNT(*) > 0
  )
ORDER BY
  message_count DESC;

-- name: GetViewChatMessges :many
SELECT
  *
FROM
  v_get_chat_messages
WHERE
  chat_id = ?;

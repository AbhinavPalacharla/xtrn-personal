/****************************************************/
/*
Models for storing chats
*/
CREATE TABLE chats (id TEXT PRIMARY KEY);

/*
HUMAN message = check messages.content

AI MESSAGE = JOIN ai_message_parts and join text_parts and tool_call_parts and prcess

TOOL MESSAGE = Check tool_call_result after joining
*/
CREATE TABLE messages (
  id TEXT PRIMARY KEY,
  role TEXT NOT NULL CHECK (
    role in ('human', 'ai', 'tool', 'system', 'xtrn_auth')
  ),
  content TEXT, -- ONLY USED FOR HUMAN MESSAGE
  stop_reason TEXT,
  chat_id TEXT NOT NULL,
  FOREIGN KEY (chat_id) REFERENCES chats (id) ON DELETE CASCADE
);

CREATE TABLE chat_auth_requests (
  id TEXT PRIMARY KEY,
  status TEXT NOT NULL CHECK (status in ('OPEN', 'COMPLETE', 'REJECTED')),
  oauth_provider_name TEXT NOT NULL,
  FOREIGN KEY (oauth_provider_name) REFERENCES oauth_providers (name) ON DELETE CASCADE,
  chat_id TEXT NOT NULL,
  FOREIGN KEY (chat_id) REFERENCES chats (id) ON DELETE CASCADE
);

CREATE TABLE ai_message_parts (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  type TEXT NOT NULL CHECK (type IN ('text', 'function')),
  part_index INTEGER NOT NULL, -- Might be able to just sort on id because AUTOINCREMENT
  message_id TEXT NOT NULL,
  FOREIGN KEY (message_id) REFERENCES messages (id) ON DELETE CASCADE
);

CREATE TABLE text_part (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  text TEXT,
  message_part_id INTEGER NOT NULL,
  FOREIGN KEY (message_part_id) REFERENCES ai_message_parts (id) ON DELETE CASCADE
);

CREATE TABLE tool_call_part (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  tool_call_id TEXT NOT NULL,
  name TEXT NOT NULL,
  arguments TEXT NOT NULL,
  message_part_id INTEGER NOT NULL,
  FOREIGN KEY (message_part_id) REFERENCES ai_message_parts (id) ON DELETE CASCADE
);

CREATE TABLE tool_call_result (
  message_id TEXT PRIMARY KEY,
  tool_call_id TEXT NOT NULL,
  name TEXT NOT NULL,
  content TEXT NOT NULL,
  is_error BOOLEAN DEFAULT FALSE NOT NULL,
  FOREIGN KEY (message_id) REFERENCES messages (id) ON DELETE CASCADE
);

CREATE VIEW v_get_chat_messages AS
SELECT
  m.id,
  m.role,
  m.content,
  m.stop_reason,
  m.chat_id,
  /* ai_message as JSON array */
  CASE
    WHEN m.role = 'ai' THEN COALESCE(
      (
        SELECT
          json_group_array(json(p.part_json))
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
                  'function',
                  'index',
                  amp.part_index,
                  'tool_call_id',
                  tcp.tool_call_id,
                  'name',
                  tcp.name,
                  'arguments',
                  json(tcp.arguments)
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
      ),
      '[]'
    )
  END AS ai_message,
  /* tool_result as JSON object */
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
  END AS tool_result,
  /* auth_request as JSON object for xtrn_auth messages */
  CASE
    WHEN m.role = 'xtrn_auth' THEN COALESCE((
      SELECT
        json_object(
          'id',
          ar.id,
          'status',
          ar.status,
          'oauth_provider_name',
          ar.oauth_provider_name,
          'provider_info',
          json_object(
            'name',
            op.name,
            'client_id',
            op.client_id,
            'callback_url',
            op.callback_url,
            'scopes',
            op.scopes
          )
        )
      FROM
        chat_auth_requests ar
        LEFT JOIN oauth_providers op ON ar.oauth_provider_name = op.name
      WHERE
        ar.chat_id = m.chat_id
        AND ar.id = (
          SELECT
            id
          FROM
            chat_auth_requests
          WHERE
            chat_id = m.chat_id
          ORDER BY
            id DESC
          LIMIT
            1
        )
    ), NULL)
  END AS auth_request
FROM
  messages m;

-- New view to get chat with auth requests and messages
CREATE VIEW v_get_chat_with_auth_and_messages AS
SELECT
  c.id AS chat_id,
  -- Aggregate auth requests as JSON array with provider info
  COALESCE(
    (
      SELECT
        json_group_array(
          json_object(
            'id',
            car.id,
            'status',
            car.status,
            'oauth_provider_name',
            car.oauth_provider_name,
            'provider_info',
            json_object(
              'name',
              op.name,
              'client_id',
              op.client_id,
              'callback_url',
              op.callback_url,
              'scopes',
              op.scopes
            )
          )
        )
      FROM
        chat_auth_requests car
        LEFT JOIN oauth_providers op ON car.oauth_provider_name = op.name
      WHERE
        car.chat_id = c.id
    ),
    '[]'
  ) AS auth_requests,
  -- Aggregate messages from v_get_chat_messages as JSON array
  COALESCE(
    (
      SELECT
        json_group_array(
          json_object(
            'id',
            v.id,
            'role',
            v.role,
            'content',
            v.content,
            'stop_reason',
            v.stop_reason,
            'ai_message',
            v.ai_message,
            'tool_result',
            v.tool_result
          )
        )
      FROM
        v_get_chat_messages v
      WHERE
        v.chat_id = c.id
    ),
    '[]'
  ) AS messages
FROM
  chats c;

/****************************************************/
/*
Possible Oauth providers
*/
CREATE TABLE oauth_providers (
  name TEXT PRIMARY KEY,
  client_id TEXT NOT NULL,
  client_secret TEXT NOT NULL,
  callback_url TEXT NOT NULL,
  scopes TEXT
);

/*
User token attached to an oauth provider
*/
CREATE TABLE oauth_tokens (
  id TEXT PRIMARY KEY,
  refresh_token TEXT UNIQUE NOT NULL,
  oauth_provider TEXT NOT NULL,
  FOREIGN KEY (oauth_provider) REFERENCES oauth_providers (name)
);

/***************************************************/
/*
mcp_servers is a list of possible MCP servers that can exist.
Must create an instance to use the server
*/
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

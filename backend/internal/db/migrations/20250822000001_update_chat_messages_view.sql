-- +goose Up
-- +goose StatementBegin

-- Drop and recreate the v_get_chat_messages view to include auth request data
DROP VIEW IF EXISTS v_get_chat_messages;

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
    WHEN m.role = 'xtrn_auth' THEN (
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
          SELECT id FROM chat_auth_requests
          WHERE chat_id = m.chat_id
          ORDER BY id DESC LIMIT 1
        )
    )
  END AS auth_request
FROM
  messages m;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Drop and recreate the old view without auth_request
DROP VIEW IF EXISTS v_get_chat_messages;

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
  END AS tool_result
FROM
  messages m;

-- +goose StatementEnd
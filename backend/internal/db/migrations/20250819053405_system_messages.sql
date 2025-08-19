-- +goose Up
-- +goose StatementBegin
PRAGMA foreign_keys = off;

-- Drop dependent view first
DROP VIEW IF EXISTS v_get_chat_messages;

-- Recreate messages table with new role constraint
CREATE TABLE messages_new (
  id TEXT PRIMARY KEY,
  role TEXT NOT NULL CHECK (role in ('human', 'ai', 'tool', 'system')),
  content TEXT, -- ONLY USED FOR HUMAN MESSAGE
  stop_reason TEXT,
  chat_id TEXT NOT NULL,
  FOREIGN KEY (chat_id) REFERENCES chats (id) ON DELETE CASCADE
);

INSERT INTO
  messages_new (id, role, content, stop_reason, chat_id)
SELECT
  id,
  role,
  content,
  stop_reason,
  chat_id
FROM
  messages;

DROP TABLE messages;

ALTER TABLE messages_new
RENAME TO messages;

-- Recreate the view to depend on the new messages table
CREATE VIEW v_get_chat_messages AS
SELECT
  m.id,
  m.role,
  m.content,
  m.stop_reason,
  m.chat_id,
  CASE
    WHEN m.role = 'ai' THEN COALESCE(
      (
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
                  'function',
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
      ),
      '[]'
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
  messages m;

PRAGMA foreign_keys = on;

-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
PRAGMA foreign_keys = off;

-- Drop dependent view first
DROP VIEW IF EXISTS v_get_chat_messages;

-- Revert messages table to old constraint
CREATE TABLE messages_old (
  id TEXT PRIMARY KEY,
  role TEXT NOT NULL CHECK (role in ('human', 'ai', 'tool')),
  content TEXT, -- ONLY USED FOR HUMAN MESSAGE
  stop_reason TEXT,
  chat_id TEXT NOT NULL,
  FOREIGN KEY (chat_id) REFERENCES chats (id) ON DELETE CASCADE
);

INSERT INTO
  messages_old (id, role, content, stop_reason, chat_id)
SELECT
  id,
  role,
  content,
  stop_reason,
  chat_id
FROM
  messages;

DROP TABLE messages;

ALTER TABLE messages_old
RENAME TO messages;

-- Recreate old view
CREATE VIEW v_get_chat_messages AS
SELECT
  m.id,
  m.role,
  m.content,
  m.stop_reason,
  m.chat_id,
  CASE
    WHEN m.role = 'ai' THEN COALESCE(
      (
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
                  'function',
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
      ),
      '[]'
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
  messages m;

PRAGMA foreign_keys = on;

-- +goose StatementEnd

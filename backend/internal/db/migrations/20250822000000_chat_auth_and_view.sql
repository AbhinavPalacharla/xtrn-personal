-- +goose Up
-- +goose StatementBegin

-- Drop existing view and table to recreate with correct schema
DROP VIEW IF EXISTS v_get_chat_with_auth_and_messages;
DROP TABLE IF EXISTS chat_auth_requests;

-- Recreate chat_auth_requests table with correct column name
CREATE TABLE chat_auth_requests (
  id TEXT PRIMARY KEY,
  status TEXT NOT NULL CHECK (status in ('OPEN', 'COMPLETE', 'REJECTED')),
  oauth_provider_name TEXT NOT NULL,
  chat_id TEXT NOT NULL,
  FOREIGN KEY (oauth_provider_name) REFERENCES oauth_providers (name) ON DELETE CASCADE,
  FOREIGN KEY (chat_id) REFERENCES chats (id) ON DELETE CASCADE
);

-- Recreate the view that includes auth requests and messages with provider info
CREATE VIEW v_get_chat_with_auth_and_messages AS
SELECT
    c.id AS chat_id,
    -- Aggregate auth requests as JSON array with provider info
    COALESCE(
        (
            SELECT json_group_array(
                json_object(
                    'id', car.id,
                    'status', car.status,
                    'oauth_provider_name', car.oauth_provider_name,
                    'provider_info', json_object(
                        'name', op.name,
                        'client_id', op.client_id,
                        'callback_url', op.callback_url,
                        'scopes', op.scopes
                    )
                )
            )
            FROM chat_auth_requests car
            LEFT JOIN oauth_providers op ON car.oauth_provider_name = op.name
            WHERE car.chat_id = c.id
        ),
        '[]'
    ) AS auth_requests,
    -- Aggregate messages from v_get_chat_messages as JSON array
    COALESCE(
        (
            SELECT json_group_array(
                json_object(
                    'id', v.id,
                    'role', v.role,
                    'content', v.content,
                    'stop_reason', v.stop_reason,
                    'ai_message', v.ai_message,
                    'tool_result', v.tool_result
                )
            )
            FROM v_get_chat_messages v
            WHERE v.chat_id = c.id
        ),
        '[]'
    ) AS messages
FROM chats c;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Drop the view and table (rollback to previous state)
DROP VIEW IF EXISTS v_get_chat_with_auth_and_messages;
DROP TABLE IF EXISTS chat_auth_requests;

-- Recreate with old schema (oauth_provider_id instead of oauth_provider_name)
CREATE TABLE chat_auth_requests (
  id TEXT PRIMARY KEY,
  status TEXT NOT NULL CHECK (status in ('OPEN', 'COMPLETE', 'REJECTED')),
  oauth_provider_id TEXT NOT NULL,
  chat_id TEXT NOT NULL,
  FOREIGN KEY (oauth_provider_id) REFERENCES oauth_providers (name) ON DELETE CASCADE,
  FOREIGN KEY (chat_id) REFERENCES chats (id) ON DELETE CASCADE
);

-- Recreate the old view without provider info
CREATE VIEW v_get_chat_with_auth_and_messages AS
SELECT
    c.id AS chat_id,
    -- Aggregate auth requests as JSON array (old format)
    COALESCE(
        (
            SELECT json_group_array(
                json_object(
                    'id', car.id,
                    'status', car.status,
                    'oauth_provider_id', car.oauth_provider_id
                )
            )
            FROM chat_auth_requests car
            WHERE car.chat_id = c.id
        ),
        '[]'
    ) AS auth_requests,
    -- Aggregate messages from v_get_chat_messages as JSON array
    COALESCE(
        (
            SELECT json_group_array(
                json_object(
                    'id', v.id,
                    'role', v.role,
                    'content', v.content,
                    'stop_reason', v.stop_reason,
                    'ai_message', v.ai_message,
                    'tool_result', v.tool_result
                )
            )
            FROM v_get_chat_messages v
            WHERE v.chat_id = c.id
        ),
        '[]'
    ) AS messages
FROM chats c;

-- +goose StatementEnd
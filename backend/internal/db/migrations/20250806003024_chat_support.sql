-- +goose Up
-- +goose StatementBegin
CREATE TABLE chats (id TEXT PRIMARY KEY);

CREATE TABLE messages (
  id TEXT PRIMARY KEY,
  role TEXT NOT NULL,
  content TEXT, -- NOT USED FOR TOOL CALL TYPE MESSAGES
  stop_reason TEXT,
  chat_id TEXT NOT NULL,
  FOREIGN KEY (chat_id) REFERENCES chats (id)
);

CREATE TABLE tool_call_request (
  message_id TEXT PRIMARY KEY,
  tool_call_id TEXT NOT NULL,
  name TEXT NOT NULL,
  arguments TEXT NOT NULL,
  FOREIGN KEY (message_id) REFERENCES messages (id)
);

CREATE TABLE tool_call_result (
  message_id TEXT PRIMARY KEY,
  tool_call_id TEXT NOT NULL,
  name TEXT NOT NULL,
  content TEXT NOT NULL,
  is_error BOOLEAN DEFAULT FALSE NOT NULL,
  FOREIGN KEY (message_id) REFERENCES messages (id)
);

-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS chats;

DROP TABLE IF EXISTS messages;

DROP TABLE IF EXISTS tool_call_request;

DROP TABLE IF EXISTS tool_call_result;

-- +goose StatementEnd

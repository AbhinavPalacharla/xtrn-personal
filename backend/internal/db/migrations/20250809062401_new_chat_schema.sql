-- +goose Up
-- +goose StatementBegin
DROP TABLE IF EXISTS tool_call_request;

DROP TABLE IF EXISTS tool_call_result;

DROP TABLE IF EXISTS messages;

DROP TABLE IF EXISTS chats;

CREATE TABLE chats (id TEXT PRIMARY KEY);

CREATE TABLE messages (
  id TEXT PRIMARY KEY,
  role TEXT NOT NULL CHECK (role in ('human', 'ai', 'tool')),
  content TEXT, -- ONLY USED FOR HUMAN MESSAGE
  stop_reason TEXT,
  chat_id TEXT NOT NULL,
  FOREIGN KEY (chat_id) REFERENCES chats (id) ON DELETE CASCADE
);

CREATE TABLE ai_message_parts (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  type TEXT CHECK (type IN ('text', 'function')),
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

-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS tool_call_result;

DROP TABLE IF EXISTS tool_call_part;

DROP TABLE IF EXISTS text_part;

DROP TABLE IF EXISTS ai_message_parts;

DROP TABLE IF EXISTS messages;

DROP TABLE IF EXISTS chats;

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

/****************************************************/
/*
Models for storing chats
*/
-- CREATE TABLE chats (
--   id TEXT PRIMARY KEY,
--   title TEXT,
--   created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
-- );
-- CREATE TABLE messages (
--   id TEXT PRIMARY KEY,
--   chat_id TEXT NOT NULL,
--   role TEXT NOT NULL CHECK (role IN ('USER', 'ASSISTANT')),
--   type TEXT NOT NULL CHECK (
--     type IN ('TEXT', 'TOOL_CALL_REQ', 'TOOL_CALL_RES')
--   ),
--   content TEXT,
--   created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
--   FOREIGN KEY (chat_id) REFERENCES chats (id)
-- );
-- CREATE TABLE tool_call_req (
--   message_id PRIMARY KEY,
--   tool_use_id TEXT NOT NULL, -- claude tool use ID
--   name TEXT NOT NULL,
--   arguments TEXT NOT NULL,
--   FOREIGN KEY (message_id) REFERENCES messages (id)
-- );
-- CREATE TABLE tool_call_res (
--   message_id PRIMARY KEY,
--   tool_use_id TEXT NOT NULL, -- claude tool use ID
--   content TEXT NOT NULL,
--   is_error BOOLEAN DEFAULT FALSE,
--   FOREIGN KEY (message_id) REFERENCES messages (id)
-- );
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

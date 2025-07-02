package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
)

type EnvSchema map[string]string

func (s EnvSchema) Value() (driver.Value, error) {
	return json.Marshal(s)
}

func (s *EnvSchema) Scan(value any) error {
	b, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("expected []byte for JSONMap, got %T", value)
	}
	return json.Unmarshal(b, s)
}

var InvalidMCPTypeError = errors.New("Invalid server type")

type MCPServerType string

const (
	MCPServerTypePublic             MCPServerType = "PUBLIC"
	MCPServerTypeAuthenticatedOauth MCPServerType = "AUTHENTICATED_OAUTH"
	MCPServerTypeAuthenticated      MCPServerType = "AUTHENTICATED"
)

func (s MCPServerType) IsValid() bool {
	switch s {
	case MCPServerTypePublic, MCPServerTypeAuthenticated, MCPServerTypeAuthenticatedOauth:
		return true
	}
	return false
}

func (s MCPServerType) MarshalJSON() ([]byte, error) {
	if !s.IsValid() {
		return nil, InvalidMCPTypeError
	}
	return json.Marshal(string(s))
}

func (s *MCPServerType) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}
	tmp := MCPServerType(str)
	if !tmp.IsValid() {
		return fmt.Errorf("%w: %s", InvalidMCPTypeError, str)
	}
	*s = tmp
	return nil
}

/********* MessageRole enum ********/

var InvalidMessageRoleError = errors.New("Invalid message role")

type MessageRole string

const (
	MessageRoleUser      MessageRole = "USER"
	MessageRoleAssistant MessageRole = "ASSISTANT"
)

func (r MessageRole) IsValid() bool {
	switch r {
	case MessageRoleUser, MessageRoleAssistant:
		return true
	}
	return false
}

func (r MessageRole) MarshalJSON() ([]byte, error) {
	if !r.IsValid() {
		return nil, InvalidMessageRoleError
	}
	return json.Marshal(string(r))
}

func (r *MessageRole) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}
	tmp := MessageRole(str)
	if !tmp.IsValid() {
		return fmt.Errorf("%w: %s", InvalidMessageRoleError, str)
	}
	*r = tmp
	return nil
}

// ---------- MessageType ----------

var InvalidMessageTypeError = errors.New("Invalid message type")

type MessageType string

const (
	MessageTypeText        MessageType = "TEXT"
	MessageTypeToolCallReq MessageType = "TOOL_CALL_REQ"
	MessageTypeToolCallRes MessageType = "TOOL_CALL_RES"
)

func (t MessageType) IsValid() bool {
	switch t {
	case MessageTypeText, MessageTypeToolCallReq, MessageTypeToolCallRes:
		return true
	}
	return false
}

func (t MessageType) MarshalJSON() ([]byte, error) {
	if !t.IsValid() {
		return nil, InvalidMessageTypeError
	}
	return json.Marshal(string(t))
}

func (t *MessageType) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}
	tmp := MessageType(str)
	if !tmp.IsValid() {
		return fmt.Errorf("%w: %s", InvalidMessageTypeError, str)
	}
	*t = tmp
	return nil
}

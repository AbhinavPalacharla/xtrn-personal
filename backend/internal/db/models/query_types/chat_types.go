package query_types

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

type OauthProviderInfo struct {
	Name        string  `json:"name"`
	ClientID    string  `json:"client_id"`
	CallbackURL string  `json:"callback_url"`
	Scopes      *string `json:"scopes,omitempty"`
}

type AuthRequest struct {
	ID                string            `json:"id"`
	Status            string            `json:"status"`
	OauthProviderName string            `json:"oauth_provider_name"`
	ProviderInfo      OauthProviderInfo `json:"provider_info"`
}

// Scan implements sql.Scanner for a JSON TEXT column -> AuthRequest
func (a *AuthRequest) Scan(value any) error {
	switch v := value.(type) {
	case nil:
		*a = AuthRequest{}
		return nil
	case []byte:
		if len(v) == 0 {
			*a = AuthRequest{}
			return nil
		}
		return json.Unmarshal(v, a)
	case string:
		if v == "" {
			*a = AuthRequest{}
			return nil
		}
		return json.Unmarshal([]byte(v), a)
	default:
		return fmt.Errorf("AuthRequest.Scan: unsupported src type %T", v)
	}
}

// Value implements driver.Valuer so you can INSERT/UPDATE this field if needed.
func (a AuthRequest) Value() (driver.Value, error) {
	b, err := json.Marshal(a)
	if err != nil {
		return nil, err
	}
	return string(b), nil
}

type AuthRequests []AuthRequest

// Scan implements sql.Scanner for a JSON TEXT column -> []AuthRequest
func (a *AuthRequests) Scan(value any) error {
	switch v := value.(type) {
	case nil:
		*a = nil
		return nil
	case []byte:
		if len(v) == 0 {
			*a = nil
			return nil
		}
		return json.Unmarshal(v, a)
	case string:
		if v == "" {
			*a = nil
			return nil
		}
		return json.Unmarshal([]byte(v), a)
	default:
		return fmt.Errorf("AuthRequests.Scan: unsupported src type %T", v)
	}
}

// Value implements driver.Valuer so you can INSERT/UPDATE this field if needed.
func (a AuthRequests) Value() (driver.Value, error) {
	if a == nil {
		return "[]", nil
	}
	b, err := json.Marshal(a)
	if err != nil {
		return nil, err
	}
	return string(b), nil
}

type ChatMessage struct {
	ID          string       `json:"id"`
	Role        string       `json:"role"`
	Content     *string      `json:"content,omitempty"`
	StopReason  *string      `json:"stop_reason,omitempty"`
	ChatID      string       `json:"chat_id"`
	AIMessage   AIParts      `json:"ai_message"`
	ToolResult  *ToolResult  `json:"tool_result,omitempty"`
	AuthRequest *AuthRequest `json:"auth_request,omitempty"`
}

type ChatMessages []ChatMessage

// Scan implements sql.Scanner for a JSON TEXT column -> []ChatMessage
func (c *ChatMessages) Scan(value any) error {
	switch v := value.(type) {
	case nil:
		*c = nil
		return nil
	case []byte:
		if len(v) == 0 {
			*c = nil
			return nil
		}
		return json.Unmarshal(v, c)
	case string:
		if v == "" {
			*c = nil
			return nil
		}
		return json.Unmarshal([]byte(v), c)
	default:
		return fmt.Errorf("ChatMessages.Scan: unsupported src type %T", v)
	}
}

// Value implements driver.Valuer so you can INSERT/UPDATE this field if needed.
func (c ChatMessages) Value() (driver.Value, error) {
	if c == nil {
		return "[]", nil
	}
	b, err := json.Marshal(c)
	if err != nil {
		return nil, err
	}
	return string(b), nil
}

type ChatWithAuthAndMessages struct {
	ChatID       string       `json:"chat_id"`
	AuthRequests AuthRequests `json:"auth_requests"`
	Messages     ChatMessages `json:"messages"`
}

type AIPart struct {
	Type       string          `json:"type"`
	Index      int             `json:"index"`
	Text       *string         `json:"text,omitempty"`
	ToolCallID string          `json:"tool_call_id,omitempty"`
	Name       string          `json:"name,omitempty"`
	Arguments  json.RawMessage `json:"arguments,omitempty"`
}

type AIParts []AIPart

// Scan implements sql.Scanner for a JSON TEXT column -> []AIPart
func (a *AIParts) Scan(value any) error {
	switch v := value.(type) {
	case nil:
		*a = nil
		return nil
	case []byte:
		if len(v) == 0 {
			*a = nil
			return nil
		}
		return json.Unmarshal(v, a)
	case string:
		if v == "" {
			*a = nil
			return nil
		}
		return json.Unmarshal([]byte(v), a)
	default:
		return fmt.Errorf("AIParts.Scan: unsupported src type %T", v)
	}
}

// Value implements driver.Valuer so you can INSERT/UPDATE this field if needed.
func (a AIParts) Value() (driver.Value, error) {
	if a == nil {
		// return JSON null or [] — pick one. [] is usually nicer.
		return "[]", nil
	}
	b, err := json.Marshal(a)
	if err != nil {
		return nil, err
	}
	return string(b), nil
}

type IntBool bool

func (b *IntBool) UnmarshalJSON(data []byte) error {
	switch string(data) {
	case "true", "1":
		*b = true
	case "false", "0":
		*b = false
	default:
		return fmt.Errorf("invalid boolean value: %s", data)
	}
	return nil
}

type ToolResult struct {
	ToolCallID string          `json:"tool_call_id"`
	Name       string          `json:"name"`
	Content    json.RawMessage `json:"content"`
	IsError    IntBool         `json:"is_error"`
}

// Scan implements sql.Scanner for a JSON TEXT column -> ToolResult
func (t *ToolResult) Scan(value any) error {
	switch v := value.(type) {
	case nil:
		*t = ToolResult{}
		return nil
	case []byte:
		if len(v) == 0 {
			*t = ToolResult{}
			return nil
		}
		return json.Unmarshal(v, t)
	case string:
		if v == "" {
			*t = ToolResult{}
			return nil
		}
		return json.Unmarshal([]byte(v), t)
	default:
		return fmt.Errorf("ToolResult.Scan: unsupported src type %T", v)
	}
}

// Optional: implement Valuer if you plan to write it back.
func (t ToolResult) Value() (driver.Value, error) {
	b, err := json.Marshal(t)
	if err != nil {
		return nil, err
	}
	return string(b), nil
}

type ToolInfo struct {
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	Schema      string  `json:"schema"`
}

type Tools []ToolInfo

// Scan implements sql.Scanner for a JSON TEXT column -> []ToolInfo
func (t *Tools) Scan(value any) error {
	switch v := value.(type) {
	case nil:
		*t = nil
		return nil
	case []byte:
		if len(v) == 0 {
			*t = nil
			return nil
		}
		return json.Unmarshal(v, t)
	case string:
		if v == "" {
			*t = nil
			return nil
		}
		return json.Unmarshal([]byte(v), t)
	default:
		return fmt.Errorf("Tools.Scan: unsupported src type %T", v)
	}
}

// Value implements driver.Valuer so you can INSERT/UPDATE this field if needed.
func (t Tools) Value() (driver.Value, error) {
	if t == nil {
		return "[]", nil
	}
	b, err := json.Marshal(t)
	if err != nil {
		return nil, err
	}
	return string(b), nil
}

type OauthInfo struct {
	ProviderName string  `json:"providerName"`
	ClientID     string  `json:"clientID"`
	ClientSecret string  `json:"clientSecret"`
	CallbackURL  string  `json:"callbackURL"`
	Scopes       *string `json:"scopes,omitempty"`
}

// Scan implements sql.Scanner for a JSON TEXT column -> OauthInfo
func (o *OauthInfo) Scan(value any) error {
	switch v := value.(type) {
	case nil:
		*o = OauthInfo{}
		return nil
	case []byte:
		if len(v) == 0 {
			*o = OauthInfo{}
			return nil
		}
		return json.Unmarshal(v, o)
	case string:
		if v == "" {
			*o = OauthInfo{}
			return nil
		}
		return json.Unmarshal([]byte(v), o)
	default:
		return fmt.Errorf("OauthInfo.Scan: unsupported src type %T", v)
	}
}

// Value implements driver.Valuer so you can INSERT/UPDATE this field if needed.
func (o OauthInfo) Value() (driver.Value, error) {
	b, err := json.Marshal(o)
	if err != nil {
		return nil, err
	}
	return string(b), nil
}

type OauthInstanceWithTools struct {
	InstanceID string     `json:"instance_id"`
	Slug       string     `json:"slug"`
	Address    string     `json:"address"`
	Tools      Tools      `json:"tools"`
	Oauth      *OauthInfo `json:"oauth,omitempty"`
}

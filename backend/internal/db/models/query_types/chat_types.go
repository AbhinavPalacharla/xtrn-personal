package query_types

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

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

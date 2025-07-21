package models

import (
	"database/sql/driver"
	"encoding/json"
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

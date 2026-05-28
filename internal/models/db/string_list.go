package db

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

type StringList []string

func (s StringList) Value() (driver.Value, error) {
	if s == nil {
		return "[]", nil
	}

	data, err := json.Marshal([]string(s))
	if err != nil {
		return nil, fmt.Errorf("marshal string list: %w", err)
	}

	return string(data), nil
}

func (s *StringList) Scan(value any) error {
	if value == nil {
		*s = nil
		return nil
	}

	var data []byte
	switch typed := value.(type) {
	case []byte:
		data = typed
	case string:
		data = []byte(typed)
	default:
		return fmt.Errorf("scan string list from %T", value)
	}

	var tags []string
	if err := json.Unmarshal(data, &tags); err != nil {
		return fmt.Errorf("unmarshal string list: %w", err)
	}

	*s = tags
	return nil
}

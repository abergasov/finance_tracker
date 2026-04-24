package types

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

type Int64Map map[string]int64

func (m *Int64Map) Scan(src any) error {
	if src == nil {
		*m = nil
		return nil
	}

	var b []byte
	switch v := src.(type) {
	case []byte:
		b = v
	case string:
		b = []byte(v)
	default:
		return fmt.Errorf("Int64Map: unsupported Scan type %T", src)
	}

	return json.Unmarshal(b, m)
}

func (m Int64Map) Value() (driver.Value, error) {
	if m == nil {
		return []byte(`{}`), nil
	}

	b, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}

	return b, nil
}

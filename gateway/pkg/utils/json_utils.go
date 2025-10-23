package utils

import (
	"encoding/json"
	"fmt"
)

func ToRawMessage(v interface{}) (json.RawMessage, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal struct to JSON: %w", err)
	}
	return json.RawMessage(data), nil
}

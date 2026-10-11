package searchmedia

import (
	"bytes"
	"encoding/json"
)

func isNull(raw json.RawMessage) bool {
	trimmed := bytes.TrimSpace(raw)

	return len(trimmed) == 0 || bytes.Equal(trimmed, []byte("null"))
}

func isObject(raw json.RawMessage) bool {
	trimmed := bytes.TrimSpace(raw)

	return len(trimmed) > 0 && trimmed[0] == '{' && json.Valid(trimmed)
}

// Nil when raw is not an object or has no such key.
func field(raw json.RawMessage, key string) json.RawMessage {
	if !isObject(raw) {
		return nil
	}

	var object map[string]json.RawMessage
	if json.Unmarshal(raw, &object) != nil {
		return nil
	}

	return object[key]
}

// Nil when something on the path is missing or is not an object.
func nested(raw json.RawMessage, path ...string) json.RawMessage {
	for _, key := range path {
		raw = field(raw, key)
		if raw == nil {
			return nil
		}
	}

	return raw
}

// Returns JSON array elements or object values as []json.RawMessage.
// Object keys are intentionally discarded.
func iterableItems(raw json.RawMessage) []json.RawMessage {
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 {
		return nil
	}

	switch trimmed[0] {
	case '[':
		var list []json.RawMessage
		if json.Unmarshal(trimmed, &list) != nil {
			return nil
		}

		return list
	case '{':
		return objectValues(trimmed)
	}

	return nil
}

func objectValues(raw json.RawMessage) []json.RawMessage {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	if _, err := decoder.Token(); err != nil {
		return nil
	}

	var values []json.RawMessage

	for decoder.More() {
		if _, err := decoder.Token(); err != nil {
			return nil
		}

		var value json.RawMessage
		if err := decoder.Decode(&value); err != nil {
			return nil
		}

		values = append(values, value)
	}

	return values
}

func nonEmptyString(raw json.RawMessage) (string, bool) {
	if isNull(raw) {
		return "", false
	}

	var value string
	if json.Unmarshal(raw, &value) != nil || value == "" {
		return "", false
	}

	return value, true
}

// "" for null or missing; strings as they are; any other value as its JSON text.
func stringOrEmpty(raw json.RawMessage) string {
	if isNull(raw) {
		return ""
	}

	var value string
	if json.Unmarshal(raw, &value) == nil {
		return value
	}

	var compact bytes.Buffer
	if json.Compact(&compact, raw) != nil {
		return ""
	}

	return compact.String()
}

// Keeps the original value and its type; "" for null or missing.
func valueOrEmpty(raw json.RawMessage) any {
	if isNull(raw) {
		return ""
	}

	return decodeValue(raw)
}

// Keeps the original value and its type; null for null or missing.
func valueOrNull(raw json.RawMessage) any {
	if isNull(raw) {
		return nil
	}

	return decodeValue(raw)
}

// Numbers stay json.Number, so they are written back exactly as they came.
func decodeValue(raw json.RawMessage) any {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	// To avoid on decoding JSON when number -> float64 by default
	decoder.UseNumber()

	var value any
	if decoder.Decode(&value) != nil {
		return nil
	}

	return value
}

func decodedValues(raws []json.RawMessage) []any {
	values := make([]any, 0, len(raws))

	for _, raw := range raws {
		values = append(values, valueOrNull(raw))
	}

	return values
}

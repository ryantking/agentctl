// Package config provides configuration file management and merging utilities.
package config

import (
	"encoding/json"
	"reflect"
)

// Merge performs a deep merge of settings with intelligent array handling.
// Strategy:
// - Nested maps: Recursive merge
// - Arrays: Union (deduplicate simple types)
// - Scalars: Overlay takes precedence
func Merge(base, overlay map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range base {
		result[k] = v
	}

	for key, value := range overlay {
		existing, exists := result[key]
		switch {
		case !exists:
			// New key - add it
			result[key] = value
		case isMap(value) && isMap(existing):
			// Both maps - recursive merge
			result[key] = Merge(
				existing.(map[string]interface{}),
				value.(map[string]interface{}),
			)
		case isSlice(value) && isSlice(existing):
			// Both slices - merge with deduplication
			result[key] = mergeLists(existing.([]interface{}), value.([]interface{}))
		default:
			// Scalar or type mismatch - overlay wins
			result[key] = value
		}
	}

	return result
}

func mergeLists(base, overlay []interface{}) []interface{} {
	result := make([]interface{}, len(base))
	copy(result, base)

	for _, item := range overlay {
		// For simple types, deduplicate
		if isSimpleType(item) {
			found := false
			for _, existing := range result {
				if reflect.DeepEqual(existing, item) {
					found = true
					break
				}
			}
			if !found {
				result = append(result, item)
			}
		} else {
			// For complex types, just append
			result = append(result, item)
		}
	}

	return result
}

func isMap(v interface{}) bool {
	_, ok := v.(map[string]interface{})
	return ok
}

func isSlice(v interface{}) bool {
	return reflect.TypeOf(v).Kind() == reflect.Slice
}

func isSimpleType(v interface{}) bool {
	switch v.(type) {
	case string, int, int64, float64, bool, nil:
		return true
	default:
		return false
	}
}

// LoadJSON loads JSON settings from bytes.
func LoadJSON(data []byte) (map[string]interface{}, error) {
	var settings map[string]interface{}
	if err := json.Unmarshal(data, &settings); err != nil {
		return nil, err
	}
	return settings, nil
}

// SaveJSON saves JSON settings to bytes with indentation.
func SaveJSON(settings map[string]interface{}) ([]byte, error) {
	return json.MarshalIndent(settings, "", "  ")
}

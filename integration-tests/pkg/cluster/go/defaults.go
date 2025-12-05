package cluster

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
)

var (
	rawSchemaOnce sync.Once
	rawSchema     map[string]any
	rawSchemaErr  error
)

func loadRawSchema() (map[string]any, error) {
	rawSchemaOnce.Do(func() {
		_, file, _, ok := runtime.Caller(0)
		if !ok {
			rawSchemaErr = fmt.Errorf("cannot resolve schema path")
			return
		}
		schemaPath := filepath.Join(filepath.Dir(file), "..", "schema.json")

		data, err := os.ReadFile(schemaPath)
		if err != nil {
			rawSchemaErr = fmt.Errorf("load schema: %w", err)
			return
		}
		if err := json.Unmarshal(data, &rawSchema); err != nil {
			rawSchemaErr = fmt.Errorf("parse schema: %w", err)
			return
		}
	})
	return rawSchema, rawSchemaErr
}

func applySchemaDefaults(raw map[string]any) error {
	schema, err := loadRawSchema()
	if err != nil {
		return err
	}
	applyDefaultsFromDoc(schema, raw)
	return nil
}

func applyDefaultsFromDoc(schema map[string]any, value any) {
	if schema == nil || value == nil {
		return
	}

	if obj, ok := value.(map[string]any); ok {
		if props, ok := schema["properties"].(map[string]any); ok {
			for name, propRaw := range props {
				prop, ok := propRaw.(map[string]any)
				if !ok {
					continue
				}
				if _, exists := obj[name]; !exists {
					if def, ok := prop["default"]; ok {
						obj[name] = def
					}
				}
				if child, exists := obj[name]; exists {
					applyDefaultsFromDoc(prop, child)
				}
			}
		}
		return
	}

	if arr, ok := value.([]any); ok {
		switch items := schema["items"].(type) {
		case map[string]any:
			for _, elem := range arr {
				applyDefaultsFromDoc(items, elem)
			}
		case []any:
			for i, elem := range arr {
				if i < len(items) {
					if itemSchema, ok := items[i].(map[string]any); ok {
						applyDefaultsFromDoc(itemSchema, elem)
					}
				} else if len(items) > 0 {
					if itemSchema, ok := items[len(items)-1].(map[string]any); ok {
						applyDefaultsFromDoc(itemSchema, elem)
					}
				}
			}
		}
	}
}

func schemaDefault(path ...string) (any, error) {
	schema, err := loadRawSchema()
	if err != nil {
		return nil, err
	}

	var node any = schema
	for _, segment := range path {
		m, ok := node.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("expected map at segment %q", segment)
		}
		val, exists := m[segment]
		if !exists {
			return nil, fmt.Errorf("schema path missing segment %q", segment)
		}
		node = val
	}
	return node, nil
}

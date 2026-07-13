// Package schema exposes the shared cluster specification JSON Schema.
// The file is embedded so compiled test programs do not depend on source
// paths recorded by the build (e.g. trimmed module cache paths).
package schema

import _ "embed"

// JSON is the raw cluster specification schema shared across all languages.
//
//go:embed schema.json
var JSON []byte

package main

import (
	"encoding/json"
	"fmt"
	"maps"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi/pkg/v3/codegen/schema"
	"github.com/pulumi/pulumi/sdk/v3/go/common/util/contract"

	"github.com/spigell/pulumi-talos-cluster/provider/cmd/pulumi-gen-talos-cluster/resources"
	"github.com/spigell/pulumi-talos-cluster/provider/pkg/provider"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Printf("Usage: %s schema <out-dir>\n", os.Args[0])
		os.Exit(1)
	}

	language, outdir := os.Args[1], os.Args[2]
	if language != "schema" {
		fmt.Printf("Only 'schema' generation is supported. Got: %s\n", language)
		os.Exit(1)
	}

	pkgSpec := generateSchema()
	mustWritePulumiSchema(pkgSpec, outdir)
}

func generateSchema() schema.PackageSpec {
	types := make(map[string]schema.ComplexTypeSpec)
	maps.Insert(types, maps.All(resources.ClusterTypes()))
	maps.Insert(types, maps.All(resources.BasicTypes()))
	maps.Insert(types, maps.All(resources.ApplyTypes()))

	res := make(map[string]schema.ResourceSpec)
	maps.Insert(res, maps.All(resources.Cluster))
	maps.Insert(res, maps.All(resources.Apply))

	return schema.PackageSpec{
		Name:              provider.ProviderName,
		Description:       "Create and manage Talos kubernetes cluster",
		License:           "Apache-2.0",
		Keywords:          []string{"pulumi", "talos", "category/infrastructure", "kind/component", "kubernetes"},
		Publisher:         "spigell",
		Repository:        fmt.Sprintf("https://github.com/spigell/pulumi-%s", provider.ProviderName),
		PluginDownloadURL: fmt.Sprintf("github://api.github.com/spigell/pulumi-%s", provider.ProviderName),
		Types:             types,
		Resources:         res,
		Language: map[string]schema.RawMessage{
			"csharp": rawMessage(map[string]any{
				"packageReferences": map[string]string{
					"Pulumi":         "3.*",
					"Pulumi.Command": "1.1.3",
				},
			}),
			"python": rawMessage(map[string]any{
				"requires": map[string]string{
					"pulumi":            ">=3.210.0,<4.0.0",
					"pulumiverse-talos": ">=0.6.0,<0.7.0",
					"pulumi-command":    "==1.1.3",
				},
				"usesIOClasses":                true,
				"liftSingleValueMethodReturns": true,
				"pyproject": map[string]any{
					"enabled": true,
				},
			}),
			"nodejs": rawMessage(map[string]any{
				"packageName": fmt.Sprintf("@spigell/pulumi-%s", provider.ProviderName),
				"devDependencies": map[string]any{
					"typescript":  "^4.3.5",
					"@types/node": "^20.0.0",
				},
				"dependencies": map[string]any{
					"@pulumi/pulumi":     "3.210.0",
					"@pulumi/command":    "v1.1.3",
					"@pulumiverse/talos": "v0.6.1", // aligned with Talos 1.11.5
				},
			}),
			"go": rawMessage(map[string]any{
				"generateResourceContainerTypes": true,
				"importBasePath":                 fmt.Sprintf("github.com/spigell/pulumi-%s/sdk/go/talos-cluster", provider.ProviderName),
			}),
		},
	}
}

func rawMessage(v any) schema.RawMessage {
	bytes, err := json.Marshal(v)
	contract.Assertf(err == nil, "error in marshaling json")
	return bytes
}

func mustWritePulumiSchema(pkgSpec schema.PackageSpec, outdir string) {
	schemaJSON, err := json.MarshalIndent(pkgSpec, "", "    ")
	if err != nil {
		panic(errors.Wrap(err, "marshaling Pulumi schema"))
	}
	mustWriteFile(outdir, "schema.json", schemaJSON)
}

func mustWriteFile(rootDir, filename string, contents []byte) {
	outPath := filepath.Join(rootDir, filename)
	if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
		panic(err)
	}
	if err := os.WriteFile(outPath, contents, 0o600); err != nil {
		panic(err)
	}
}

package main

import (
	"encoding/json"
	"fmt"
	"maps"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	dotnetgen "github.com/pulumi/pulumi/pkg/v3/codegen/dotnet"
	gogen "github.com/pulumi/pulumi/pkg/v3/codegen/go"
	nodejsgen "github.com/pulumi/pulumi/pkg/v3/codegen/nodejs"
	pygen "github.com/pulumi/pulumi/pkg/v3/codegen/python"

	"github.com/pulumi/pulumi/pkg/v3/codegen/schema"
	"github.com/pulumi/pulumi/sdk/v3/go/common/util/contract"

	"github.com/spigell/pulumi-talos-cluster/provider/cmd/pulumi-gen-talos-cluster/resources"
	"github.com/spigell/pulumi-talos-cluster/provider/pkg/provider"
)

const tool = "Pulumi SDK Generator"

// Language is the SDK language.
type Language string

const (
	NodeJS Language = "nodejs"
	DotNet Language = "dotnet"
	Go     Language = "go"
	Python Language = "python"
	Schema Language = "schema"
)

func main() {
	printUsage := func() {
		fmt.Printf("Usage: %s <language> <out-dir> [schema-file] [version]\n", os.Args[0])
	}

	args := os.Args[1:]
	if len(args) < 2 {
		printUsage()
		os.Exit(1)
	}

	language, outdir := Language(args[0]), args[1]

	var schemaFile string
	var version string
	if language != Schema {
		if len(args) < 4 {
			printUsage()
			os.Exit(1)
		}
		schemaFile, version = args[2], args[3]
	}

	switch language {
	case NodeJS:
		genNodeJS(readSchema(schemaFile, version), outdir)
	case DotNet:
		genDotNet(readSchema(schemaFile, version), outdir)
	case Go:
		genGo(readSchema(schemaFile, version), outdir)
	case Python:
		genPython(readSchema(schemaFile, version), outdir)
	case Schema:
		pkgSpec := generateSchema()
		mustWritePulumiSchema(pkgSpec, outdir)
	default:
		panic(fmt.Sprintf("Unrecognized language %q", language))
	}
}

func generateSchema() schema.PackageSpec {
	types := make(map[string]schema.ComplexTypeSpec)

	maps.Insert(types, maps.All(resources.ClusterTypes()))
	maps.Insert(types, maps.All(resources.BasicTypes()))
	maps.Insert(types, maps.All(resources.BootstrapTypes()))
	maps.Insert(types, maps.All(resources.ApplyTypes()))

	res := make(map[string]schema.ResourceSpec)
	maps.Insert(res, maps.All(resources.Cluster))
	maps.Insert(res, maps.All(resources.Bootstrap))
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
					"Pulumi.Command": "1.1.0-alpha.1727883369",
				},
			}),
			"python": rawMessage(map[string]any{
				"requires": map[string]string{
					"pulumi":            ">=3.0.0,<4.0.0",
					"pulumiverse-talos": "0.4.1",
					"pulumi-command":    "1.0.1",
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
					"@pulumi/pulumi":     "^3.0.0",
					"@pulumi/command":    "v1.0.1",
					"@pulumiverse/talos": "v0.4.1",
				},
			}),
			"go": rawMessage(map[string]interface{}{
				"generateResourceContainerTypes": true,
				"importBasePath":                 fmt.Sprintf("github.com/spigell/pulumi-%s/sdk/go/talos-cluster", provider.ProviderName),
			}),
		},
	}
}

func rawMessage(v interface{}) schema.RawMessage {
	bytes, err := json.Marshal(v)
	contract.Assertf(err == nil, "error in marshaling json")
	return bytes
}

func readSchema(schemaPath string, version string) *schema.Package {
	// Read in, decode, and import the schema.
	schemaBytes, err := os.ReadFile(schemaPath)
	if err != nil {
		panic(err)
	}

	var pkgSpec schema.PackageSpec
	if err = json.Unmarshal(schemaBytes, &pkgSpec); err != nil {
		panic(err)
	}
	pkgSpec.Version = version

	pkg, err := schema.ImportSpec(pkgSpec, nil)
	if err != nil {
		panic(err)
	}
	return pkg
}

func genDotNet(pkg *schema.Package, outdir string) {
	files, err := dotnetgen.GeneratePackage(tool, pkg, map[string][]byte{}, nil)
	if err != nil {
		panic(err)
	}
	mustWriteFiles(outdir, files)
}

func genGo(pkg *schema.Package, outdir string) {
	files, err := gogen.GeneratePackage(tool, pkg, nil)
	if err != nil {
		panic(err)
	}
	mustWriteFiles(outdir, files)
}

func genPython(pkg *schema.Package, outdir string) {
	files, err := pygen.GeneratePackage(tool, pkg, map[string][]byte{})
	if err != nil {
		panic(err)
	}
	mustWriteFiles(outdir, files)
}

func genNodeJS(pkg *schema.Package, outdir string) {
	files, err := nodejsgen.GeneratePackage(tool, pkg, map[string][]byte{}, nil, false)
	if err != nil {
		panic(err)
	}
	mustWriteFiles(outdir, files)
}

func mustWriteFiles(rootDir string, files map[string][]byte) {
	for filename, contents := range files {
		mustWriteFile(rootDir, filename, contents)
	}
}

func mustWriteFile(rootDir, filename string, contents []byte) {
	outPath := filepath.Join(rootDir, filename)

	if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
		panic(err)
	}
	err := os.WriteFile(outPath, contents, 0o600)
	if err != nil {
		panic(err)
	}
}

func mustWritePulumiSchema(pkgSpec schema.PackageSpec, outdir string) {
	schemaJSON, err := json.MarshalIndent(pkgSpec, "", "    ")
	if err != nil {
		panic(errors.Wrap(err, "marshaling Pulumi schema"))
	}
	mustWriteFile(outdir, "schema.json", schemaJSON)
}

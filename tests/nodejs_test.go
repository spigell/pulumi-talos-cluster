package examples

import (
	"path/filepath"
	"testing"

	"github.com/pulumi/pulumi/pkg/v3/testing/integration"
)

func TestHcloudClusterJS(t *testing.T) {
	test := getJSBaseOptions(t).
		With(integration.ProgramTestOptions{
			RunUpdateTest: false,
			Dir:           filepath.Join(getTestPrograms(t), "hcloud-simple-js"),
			// ExtraRuntimeValidation: func(t *testing.T, info integration.RuntimeValidationStackInfo) {
		})

	programTest(t, &test)
}

func getJSBaseOptions(t *testing.T) integration.ProgramTestOptions {
	base := getBaseOptions(t)
	baseJS := base.With(integration.ProgramTestOptions{
		Dependencies: []string{
			"@spigell/pulumi-talos-cluster",
		},
		Verbose: true,
	})

	return baseJS
}

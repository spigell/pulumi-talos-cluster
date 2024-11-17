package examples

import (
	"path/filepath"
	"testing"

	"github.com/pulumi/pulumi/pkg/v3/testing/integration"
)

func TestHcloudClusterGo(t *testing.T) {
	test := getGoBaseOptions(t).
		With(integration.ProgramTestOptions{
			RunUpdateTest: false,
			Dir:           filepath.Join(getTestPrograms(t), "hcloud-simple-go"),
			// ExtraRuntimeValidation: func(t *testing.T, info integration.RuntimeValidationStackInfo) {
		})

	programTest(t, &test)
}

func getGoBaseOptions(t *testing.T) integration.ProgramTestOptions {
	base := getBaseOptions(t)
	goBase := base.With(integration.ProgramTestOptions{
		Dependencies: []string{
			// Path to GO SDK
			getCwd(t) + "../../sdk",
		},
		Verbose: true,
	})

	return goBase
}

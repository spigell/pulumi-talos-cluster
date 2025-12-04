package examples

import (
	"path/filepath"
	"testing"

	"github.com/pulumi/pulumi/pkg/v3/testing/integration"
)

func TestHcloudHAClusterPython(t *testing.T) {
	test := getPythonBaseOptions(t).With(integration.ProgramTestOptions{
		RunUpdateTest: false,
		Dir:           filepath.Join(getTestPrograms(t), "hcloud-ha-py"),
	})

	programTest(t, &test)
}

func getPythonBaseOptions(t *testing.T) integration.ProgramTestOptions {
	base := getBaseOptions(t)
	cwd := getCwd(t)

	return base.With(integration.ProgramTestOptions{
		Dependencies: []string{
			filepath.Join(cwd, "..", "sdk", "python"),
			filepath.Join(cwd), // integration-tests package (pyproject.toml)
		},
	})
}

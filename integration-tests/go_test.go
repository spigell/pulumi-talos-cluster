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
			Dir:           filepath.Join(getTestPrograms(t), "hcloud-go"),
			// EditDirs: []integration.EditDir{
			//		{
			//			Dir: filepath.Join(getTestPrograms(t), "hcloud-go/configuration"),
			//			Additive: true,
			//		},
			//	},
		})

	programTest(t, &test)
}

func TestHcloudHAClusterGo(t *testing.T) {
	test := getGoBaseOptions(t).
		With(integration.ProgramTestOptions{
			RunUpdateTest: false,
			Dir:           filepath.Join(getTestPrograms(t), "hcloud-ha-go"),
		})

	programTest(t, &test)
}

func getGoBaseOptions(t *testing.T) integration.ProgramTestOptions {
	base := getBaseOptions(t)
	goBase := base.With(integration.ProgramTestOptions{
		Dependencies: []string{
			// Path to GO SDK
			getCwd(t) + "../../sdk",
			getCwd(t) + "../../integration-tests",
		},
	})

	return goBase
}

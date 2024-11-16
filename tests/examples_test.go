package examples

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/pulumi/pulumi/pkg/v3/testing/integration"
	"github.com/stretchr/testify/assert"
)

func programTest(
	t *testing.T,
	opts *integration.ProgramTestOptions,
) {
	pt := integration.ProgramTestManualLifeCycle(t, opts)

	destroyStack := func() {
		destroyErr := pt.TestLifeCycleDestroy()

		assert.NoError(t, destroyErr)
	}

	// Inlined pt.TestLifeCycleInitAndDestroy()
	testLifeCycleInitAndDestroy := func() error {
		err := pt.TestLifeCyclePrepare()
		if err != nil {
			return fmt.Errorf("copying test to temp dir: %w", err)
		}

		pt.TestFinished = false
		defer pt.TestCleanUp()

		err = pt.TestLifeCycleInitialize()
		if err != nil {
			return fmt.Errorf("initializing test project: %w", err)
		}
		// Ensure that before we exit, we attempt to destroy and remove the stack.
		defer destroyStack()

		if err = pt.TestPreviewUpdateAndEdits(); err != nil {
			return fmt.Errorf("running test preview, update, and edits: %w", err)
		}
		pt.TestFinished = true
		return nil
	}

	err := testLifeCycleInitAndDestroy()
	if !errors.Is(err, integration.ErrTestFailed) {
		assert.NoError(t, err)
	}
}

func getBaseOptions(t *testing.T) integration.ProgramTestOptions {
	pathEnv, err := providerPluginPathEnv()
	if err != nil {
		t.Fatalf("failed to build provider plugin PATH: %v", err)
	}
	return integration.ProgramTestOptions{
		Env:                    []string{pathEnv},
		DecryptSecretsInOutput: true,
		ExpectRefreshChanges:   false,
		RetryFailedSteps:       false,
		SecretsProvider: getSecretProvider(),
	}
}

func providerPluginPathEnv() (string, error) {
	// Local build of the plugin.
	pluginDir := filepath.Join("..", "bin", "test")
	absPluginDir, err := filepath.Abs(pluginDir)
	if err != nil {
		return "", err
	}

	pathSeparator := ":"
	if runtime.GOOS == "windows" {
		pathSeparator = ";"
	}
	return "PATH=" + os.Getenv("PATH") + pathSeparator + absPluginDir, nil
}

func getCwd(t *testing.T) string {
	cwd, err := os.Getwd()
	if err != nil {
		t.FailNow()
	}

	return cwd
}

func getSecretProvider() string {
	prov := "default"

	if os.Getenv("SECRET_PROVIDER") != "" {
		prov = os.Getenv("SECRET_PROVIDER")
	}

	return prov
}
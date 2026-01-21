package applier_test

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// ProxyMock intercepts local.Command resources and executes them for real, capturing talosctl output.
type ProxyMock struct {
	pulumi.MockResourceMonitor
	t          *testing.T
	lastStdout string
	lastCreate string
	lastStderr string
	lastExit   int
}

func (m *ProxyMock) Call(args pulumi.MockCallArgs) (resource.PropertyMap, error) {
	cmdStr := args.Args["command"].StringValue()
	if m.t != nil {
		m.t.Logf("running invoke command: %s (dir=%s)", truncate(cmdStr, logLimit()), args.Args["dir"])
	}

	interp := interpreterFromPV(args.Args["interpreter"])
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, interp[0], append(interp[1:], cmdStr)...)
	if args.Args["dir"].HasValue() {
		cmd.Dir = args.Args["dir"].StringValue()
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	exitCode := -1
	if cmd.ProcessState != nil {
		exitCode = cmd.ProcessState.ExitCode()
	}
	if strings.Contains(cmdStr, "talosctl") {
		m.lastStdout = stdout.String()
		m.lastStderr = stderr.String()
		m.lastExit = exitCode
	}
	if err != nil {
		return resource.PropertyMap{
			"stdout":   resource.NewStringProperty(stdout.String()),
			"stderr":   resource.NewStringProperty(stderr.String()),
			"exitCode": resource.NewNumberProperty(float64(exitCode)),
		}, err
	}

	return resource.PropertyMap{
		"stdout":   resource.NewStringProperty(stdout.String()),
		"stderr":   resource.NewStringProperty(stderr.String()),
		"exitCode": resource.NewNumberProperty(float64(exitCode)),
	}, nil
}

func (m *ProxyMock) NewResource(args pulumi.MockResourceArgs) (string, resource.PropertyMap, error) {
	cmdStr := args.Inputs["create"].StringValue()
	m.lastCreate = cmdStr
	if m.t != nil {
		m.t.Logf("running command: %s (dir=%s)", truncate(cmdStr, logLimit()), args.Inputs["dir"])
	}
	dir := ""
	if args.Inputs["dir"].HasValue() {
		dir = args.Inputs["dir"].StringValue()
	}

	if dir != "" {
		_ = os.MkdirAll(dir, 0o700)
	}

	interp := interpreterFromPV(args.Inputs["interpreter"])

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, interp[0], append(interp[1:], cmdStr)...)
	if dir != "" {
		cmd.Dir = dir
	}

	output, err := cmd.CombinedOutput()
	if strings.Contains(cmdStr, "talosctl") {
		m.lastStdout = string(output)
		if cmd.ProcessState != nil {
			m.lastExit = cmd.ProcessState.ExitCode()
		} else {
			m.lastExit = -1
		}
	}
	if err != nil {
		return args.Name + "_id", resource.PropertyMap{
			"stdout": resource.NewStringProperty(string(output)),
			"stderr": resource.NewStringProperty(""),
		}, err
	}

	return args.Name + "_id", resource.PropertyMap{
		"stdout": resource.NewStringProperty(m.lastStdout),
		"stderr": resource.NewStringProperty(""),
	}, nil
}

package talosctl

import (
	"encoding/base64"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/pulumi/pulumi-command/sdk/go/command/local"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

const (
	talosctlConfigName = "talosctl.yaml"
)

type Talosctl struct {
	ctx          *pulumi.Context
	deps         []pulumi.Resource
	Binary       string
	BasicCommand string
	Home         *TalosctlHome
}

type TalosctlHome struct {
	Dir string
}

type TalosctlArgs struct {
	// PrepareConfig is the talosctl.yaml content to write to $TALOSCTL_HOME/talosctl.yaml.
	PrepareConfig pulumi.StringInput

	Args pulumi.StringInput
	RetryCount int
	Environment pulumi.StringMap
	Triggers pulumi.Array
	Interpreter pulumi.StringArray
}



func New(ctx *pulumi.Context, home string, deps []pulumi.Resource) *Talosctl {
	binary := "talosctl"

	return &Talosctl{
		ctx:          ctx,
		deps: deps,
		Binary:       binary,
		BasicCommand: fmt.Sprintf("%s --talosconfig %s/%s", binary, home, talosctlConfigName),
		Home: &TalosctlHome{
			Dir: home,
		},
	}
}

func (t *Talosctl) RunCommand(
	name string,
	a *TalosctlArgs,
	opts ...pulumi.ResourceOption,
) (pulumi.Resource, error) {
	if a == nil || a.Args == nil {
		return nil, fmt.Errorf("TalosctlArgs.Args is required")
	}

	if a.RetryCount == 0 {
		a.RetryCount = 5
	}

	prepared := a.PrepareConfig.ToStringOutput().ApplyT(func(s string) pulumi.BoolOutput {
		return t.prepare(s)
	}).(pulumi.BoolOutput)

	// ----- MAIN (visible) -----
	env := a.Environment
	if env == nil {
		env = pulumi.StringMap{}
	}

	// Gate the main Create on successful prepare:
	// If `prepared` is rejected (prepare failed), this Apply never yields a value
	// and the main resource will fail without executing.
	createGated := prepared.ApplyT(func(_ bool) (pulumi.StringPtrInput, error) {
		// pass through your original Create string (or StringOutput)
		return a.Args, nil
	}).(pulumi.StringOutput)

	main, err := local.NewCommand(t.ctx, name, &local.CommandArgs{
		Create:      createGated.ApplyT(func (args string) string {
			return withBashRetry(fmt.Sprintf(
				"%s %s", t.BasicCommand, args,
			), fmt.Sprint(a.RetryCount))
		}).(pulumi.StringOutput),
		Interpreter: a.Interpreter,
		Environment: env,
		Triggers:    a.Triggers,
	}, opts...)
	if err != nil {
		return nil, err
	}

	// --- CLEAN (hidden RunOutput) ---
	_ = local.RunOutput(t.ctx, local.RunOutputArgs{
	    Command: pulumi.Sprintf(`rm -rfv %q`, t.Home.Dir),
	}, pulumi.DependsOn([]pulumi.Resource{main}))

	return main, nil
}

func (t *Talosctl) prepare(config string) pulumi.BoolOutput {
	talosConfigPath := filepath.Join(t.Home.Dir, talosctlConfigName)
	encoded := base64.StdEncoding.EncodeToString([]byte(config))

	// Run in the right stage, with your deps.
	cmd := fmt.Sprintf(
		`mkdir -p %s && umask 077; printf %%s %q | base64 -d > %s && chmod 600 %s`,
		t.Home.Dir, encoded, talosConfigPath, talosConfigPath,
	)

	// local.RunOutput returns an Output of stdout (string). If the command fails,
	// this Output becomes a *rejected* Output, which is exactly how we "return an error".
	run := local.RunOutput(t.ctx, local.RunOutputArgs{
		Command: pulumi.String(cmd),
	}, pulumi.DependsOn(t.deps))

	// Map success to true; propagate any error unchanged.
	return run.Stderr().ApplyT(func(s string) (bool, error) {
		if s != "" {
			return false, fmt.Errorf("prepare error (stderr is not empty): %s", s)
		}
		return true, nil
	}).(pulumi.BoolOutput)
}

func withBashRetry(cmd string, retryCount string) string {
	return fmt.Sprintf(strings.Join([]string{
		"n=0",
		"until [ $n -ge %[1]s ]",
		"do %s && break",
		"sleep 10",
		"n=$((n+1))",
		"done",
		// Exiting with 0 if command succeeded.
		// Otherwise exit with 10 exit code.
		"[ $n -ge %[1]s ] && exit 10 || true",
	}, " ; "), retryCount, cmd)
}

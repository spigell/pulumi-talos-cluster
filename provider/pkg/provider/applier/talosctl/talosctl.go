package talosctl

import (
	"encoding/base64"
	"fmt"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/pulumi/pulumi-command/sdk/go/command/local"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

const (
	talosctlConfigName = "talosctl.yaml"
	defaultRetryCount = 5
)

var (
	interpreter = []string{
		"/bin/bash",
		"-c",
	}
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
	TalosConfig  pulumi.StringInput
	Args           pulumi.StringInput
	RetryCount     int
	Environment    pulumi.StringMap
	Triggers       pulumi.Array
	Interpreter    pulumi.StringArray
	AdditionalFiles []ExtraFile // <-- NEW
}

type ExtraFile struct {
	// Path relative to TALOS_HOME (e.g., "manifests/etcd.yaml") or absolute.
	Path    string
	Content pulumi.StringInput // file body
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
	createGated, env, err := t.prepareAndGate(a)
	if err != nil {
		return nil, err
	}

	main, err := local.NewCommand(t.ctx, name, &local.CommandArgs{
		Create: createGated.ApplyT(func(args string) string {
			return withBashRetry(
				fmt.Sprintf("%s %s", t.BasicCommand, args),
				fmt.Sprint(a.RetryCount),
			)
		}).(pulumi.StringOutput),
		Interpreter: pulumi.ToStringArray(interpreter),
		Environment: env,
		Triggers:    a.Triggers,
	}, opts...)
	if err != nil {
		return nil, err
	}

	// Hidden cleanup after the resource completes.
	_ = local.RunOutput(t.ctx, local.RunOutputArgs{
		Command: pulumi.Sprintf(`rm -rfv %q`, t.Home.Dir),
	}, pulumi.DependsOn([]pulumi.Resource{main}))

	return main, nil
}

func (t *Talosctl) RunGetCommand(
	name string,
	a *TalosctlArgs,
	deps []pulumi.Resource,
) (pulumi.StringOutput, error) {
	createGated, env, err := t.prepareAndGate(a)
	if err != nil {
		// Return a zero output with error
		return pulumi.StringOutput{}, err
	}

	// Compose main + inline cleanup (no resource to depend on)
	cmd := createGated.ApplyT(func(args string) string {
		runtime.Breakpoint()
		main := withBashRetry(
			fmt.Sprintf("%s %s", t.BasicCommand, args),
			fmt.Sprint(a.RetryCount),
		)
		cleanup := fmt.Sprintf("rm -rfv %s", t.Home.Dir)
		return main + " && " + cleanup
	}).(pulumi.StringOutput)

	out := local.RunOutput(t.ctx, local.RunOutputArgs{
		Command:     cmd.ApplyT(func (s string) string {
			runtime.Breakpoint()
			return s
		}).(pulumi.StringOutput),
		Interpreter: pulumi.ToStringArray(interpreter),
		Environment: env,
	}, pulumi.DependsOn(deps))


	return out.Stdout(), nil
}

func (t *Talosctl) prepareAndGate(a *TalosctlArgs) (createGated pulumi.StringOutput, env pulumi.StringMap, err error) {
	if a == nil || a.Args == nil {
		return pulumi.StringOutput{}, nil, fmt.Errorf("TalosctlArgs.Args is required")
	}
	if a.RetryCount == 0 {
		a.RetryCount = defaultRetryCount
	}

	env = a.Environment
	if env == nil {
		env = pulumi.StringMap{}
	}

	// Prepare: write talosctl.yaml + additional files
	prepared := t.prepareAll(a.TalosConfig, a.AdditionalFiles)

	// Gate the main args on successful prepare
	createGated = prepared.ApplyT(func(_ bool) (pulumi.StringPtrInput, error) {
		runtime.Breakpoint()
		return a.Args, nil
	}).(pulumi.StringOutput)

	return createGated, env, nil
}

// prepareAll builds one shell that writes talosctl.yaml and any AdditionalFiles.
// It uses `local.RunOutput` as the prepare step, returning a BoolOutput.
func (t *Talosctl) prepareAll(mainCfg pulumi.StringInput, extras []ExtraFile) pulumi.BoolOutput {
	// Gather inputs: main config + each extra file content
	inputs := []any{mainCfg}
	inputs = append(inputs, mainCfg)
	for _, f := range extras {
		inputs = append(inputs, f.Content)
	}

	// Build the full shell command as a StringOutput from resolved inputs
	cmd := pulumi.All(inputs...).ApplyT(func(resolved []any) string {
		// resolved[0] = main talosctl.yaml content
		main := resolved[0].(string)

		// 1) talosctl.yaml
		talosConfigPath := filepath.Join(t.Home.Dir, talosctlConfigName)
		mainB64 := base64.StdEncoding.EncodeToString([]byte(main))
		var b strings.Builder
		// Ensure TALOS_HOME exists, private perms
		fmt.Fprintf(&b, `mkdir -p %s && umask 077; `, t.Home.Dir)
		// Write talosctl.yaml
		fmt.Fprintf(&b, `printf %%s %q | base64 -d > %s && chmod 600 %s`,
			mainB64, talosConfigPath, talosConfigPath)


		// 2) Additional files
		for i, ef := range extras {
			content := resolved[1+i].(string)
			contentB64 := base64.StdEncoding.EncodeToString([]byte(content))

			// Resolve final path (absolute stays absolute, otherwise under TALOS_HOME)
			finalPath := ef.Path
			if !filepath.IsAbs(finalPath) {
				finalPath = filepath.Join(t.Home.Dir, ef.Path)
			}
			parent := filepath.Dir(finalPath)

			// Ensure parent dir exists, write, and lock down perms
			fmt.Fprintf(&b, ` && mkdir -p %s && printf %%s %q | base64 -d > %s && chmod 600 %s`,
				parent, contentB64, finalPath, finalPath)
		}

		return b.String()
	}).(pulumi.StringOutput)

	// Execute prepare (single invoke). Error propagates via rejected output.
	run := local.RunOutput(t.ctx, local.RunOutputArgs{
		Command: cmd,
	}, pulumi.DependsOn(t.deps))

	// Map stderr to success/failure, preserving errors.
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


package hooks

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// EtcdReadyHook returns a hook function that waits for the etcd cluster to become healthy.
func EtcdReadyHook(logger pulumi.Log) pulumi.ResourceHookFunction {
	return func(args *pulumi.ResourceHookArgs) error {
		const (
			maxRetries    = 15
			healthTimeout = 7 * time.Second
			listTimeout   = 7 * time.Second
			okStreak      = 2
		)

		env := args.NewInputs["environment"].ObjectValue().Mappable()

		ip, _ := env["NODE_IP"].(string)
		if ip == "" {
			return fmt.Errorf("environment.NODE_IP is missing or not a string")
		}
		workDir, _ := env["TALOSCTL_HOME"].(string)
		if workDir == "" {
			return fmt.Errorf("environment.TALOSCTL_HOME is missing or not a string")
		}
		targetStr, _ := env["ETCD_MEMBER_TARGET"].(string)
		if targetStr == "" {
			return fmt.Errorf("environment.ETCD_MEMBER_TARGET is missing or not a string")
		}

		expected, err := strconv.Atoi(targetStr)
		if err != nil {
			return fmt.Errorf("invalid environment.ETCD_MEMBER_TARGET %q: %w", targetStr, err)
		}

		talosconfigPath := filepath.Join(workDir, "talosctl.yaml")

		run := func(timeout time.Duration, args ...string) ([]byte, error) {
			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			defer cancel()

			full := append([]string{"--talosconfig", talosconfigPath}, args...)
			logger.Debug(fmt.Sprintf("exec: talosctl %s", strings.Join(full, " ")), nil)

			cmd := exec.CommandContext(ctx, "talosctl", full...)
			cmd.Dir = workDir

			out, err := cmd.CombinedOutput()
			if err != nil {
				return out, fmt.Errorf("talosctl %v failed: %w: %s", full, err, strings.TrimSpace(string(out)))
			}
			return out, nil
		}

		consecutiveOK := 0

		for attempt := 1; attempt <= maxRetries; attempt++ {
			backoff := time.Duration(attempt) * time.Second

			// 1) health/status
			if _, err := run(healthTimeout, "-n", ip, "-e", ip, "etcd", "status"); err != nil {
				logger.Debug(fmt.Sprintf("talos-cluster: etcd status attempt %d/%d failed: %v", attempt, maxRetries, err), nil)
				time.Sleep(backoff)
				continue
			}

			// 2) members (tabular)
			out, err := run(listTimeout, "-n", ip, "-e", ip, "etcd", "members")
			if err != nil {
				logger.Debug(fmt.Sprintf("talos-cluster: etcd members attempt %d/%d failed: %v", attempt, maxRetries, err), nil)
				time.Sleep(backoff)
				continue
			}

			got, perr := countEtcdMembersFromTable(out)
			if perr != nil {
				logger.Debug(fmt.Sprintf("talos-cluster: parse members attempt %d/%d failed: %v", attempt, maxRetries, perr), nil)
				time.Sleep(backoff)
				continue
			}

			if got != expected {
				consecutiveOK = 0
				logger.Debug(fmt.Sprintf("talos-cluster: attempt %d/%d: expected %d members, got %d", attempt, maxRetries, expected, got), nil)
				time.Sleep(backoff)
				continue
			}

			consecutiveOK++
			if consecutiveOK < okStreak {
				logger.Debug(fmt.Sprintf("talos-cluster: attempt %d/%d: matched (%d). waiting for stability %d/%d", attempt, maxRetries, got, consecutiveOK, okStreak), nil)
				time.Sleep(backoff / 2)
				continue
			}

			logger.Info(fmt.Sprintf("talos-cluster: etcd health check passed after attempt %d/%d. members=%d", attempt, maxRetries, got), nil)
			return nil
		}

		return fmt.Errorf("talos-cluster: etcd health check failed after %d attempts", maxRetries)
	}
}

/* An example of talosctl etcd members
NODE            ID                 HOSTNAME        PEER URLS                  CLIENT URLS                LEARNER
91.98.138.169   97f365161a13b437   talos-8me-09g   https://10.10.10.2:2380    https://10.10.10.2:2379    false
91.98.138.169   c22a6165837acd5b   talos-apx-beq   https://10.10.10.10:2380   https://10.10.10.10:2379   false
91.98.138.169   d65010d57cf22d49   talos-q2p-yyy   https://10.10.10.5:2380    https://10.10.10.5:2379    false
*/

// countEtcdMembersFromTable parses `talosctl etcd members` tabular output.
// It ignores the first non-empty header line and counts subsequent non-empty lines.
func countEtcdMembersFromTable(out []byte) (int, error) {
	sc := bufio.NewScanner(bytes.NewReader(out))
	sc.Buffer(make([]byte, 0, 64*1024), 1<<20)

	lines := make([]string, 0, 8)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line != "" {
			lines = append(lines, line)
		}
	}
	if err := sc.Err(); err != nil {
		return 0, fmt.Errorf("scan output: %w", err)
	}
	if len(lines) == 0 {
		return 0, fmt.Errorf("no output from talosctl etcd members")
	}

	// First non-empty line is the header. Everything after is a member row.
	memberRows := lines[1:]
	count := 0
	for _, row := range memberRows {
		// Be safe: skip accidental separator lines, etc.
		// Real rows should have multiple columns when split by fields.
		if len(strings.Fields(row)) >= 3 {
			count++
		}
	}

	return count, nil
}

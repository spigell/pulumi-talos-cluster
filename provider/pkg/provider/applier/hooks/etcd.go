package hooks

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/spigell/pulumi-talos-cluster/provider/pkg/provider/applier/talosctl"
)

type PeerStatus struct {
	ID      string `json:"id"`
	Learner bool   `json:"learner"`
}

// EtcdReadyHook returns a hook function that waits for the etcd cluster to become healthy.
func EtcdReadyHook(logger pulumi.Log) pulumi.ResourceHookFunction {
	return func(args *pulumi.ResourceHookArgs) error {
		const (
			maxRetries    = 30
			healthTimeout = 7 * time.Second
			listTimeout   = 7 * time.Second
			okStreak      = 2
		)

		// 1) Read env
		ip, workDir, expected, err := readEtcdEnv(args)
		if err != nil {
			return err
		}

		// 2) Build runner
		cli := talosctl.New().WithNodeIP(ip)
		run := makeTalosRunner(cli, workDir, logger)

		// 3) Wait loop with simple linear backoff
		consecutiveOK := 0
		for attempt := 1; attempt <= maxRetries; attempt++ {
			backoff := time.Duration(attempt) * time.Second

			// 3.1) health/status
			if err := checkEtcdStatus(run, healthTimeout); err != nil {
				logger.Debug(fmt.Sprintf("talos-cluster: etcd status attempt %d/%d failed: %v", attempt, maxRetries, err), nil)
				if err := sleepWithContext(args.Context, backoff); err != nil {
					return err
				}
				continue
			}

			// 3.2) members
			peers, err := listEtcdPeers(run, listTimeout)
			if err != nil {
				logger.Debug(fmt.Sprintf("talos-cluster: etcd members attempt %d/%d failed: %v", attempt, maxRetries, err), nil)
				if err := sleepWithContext(args.Context, backoff); err != nil {
					return err
				}
				continue
			}

			// 3.3) validate
			ok, reason := peersReady(peers, expected)
			if !ok {
				consecutiveOK = 0
				logger.Debug(fmt.Sprintf("talos-cluster: attempt %d/%d: %s", attempt, maxRetries, reason), nil)
				if err := sleepWithContext(args.Context, backoff); err != nil {
					return err
				}
				continue
			}

			// 3.4) stability window
			consecutiveOK++
			if consecutiveOK < okStreak {
				logger.Debug(fmt.Sprintf(
					"talos-cluster: attempt %d/%d: matched (%d). waiting for stability %d/%d",
					attempt, maxRetries, len(peers), consecutiveOK, okStreak,
				), nil)
				if err := sleepWithContext(args.Context, backoff/2); err != nil {
					return err
				}
				continue
			}

			logger.Info(fmt.Sprintf("talos-cluster: etcd health check passed after attempt %d/%d. members=%d",
				attempt, maxRetries, len(peers)), nil)
			return nil
		}

		return fmt.Errorf("talos-cluster: etcd health check failed after %d attempts", maxRetries)
	}
}

func readEtcdEnv(args *pulumi.ResourceHookArgs) (ip, workDir string, expected int, err error) {
	env := args.NewInputs["environment"].ObjectValue().Mappable()

	ip, _ = env["NODE_IP"].(string)
	if ip == "" {
		return "", "", 0, fmt.Errorf("environment.NODE_IP is missing or not a string")
	}
	workDir, _ = env["TALOSCTL_HOME"].(string)
	if workDir == "" {
		return "", "", 0, fmt.Errorf("environment.TALOSCTL_HOME is missing or not a string")
	}
	targetStr, _ := env["ETCD_MEMBER_TARGET"].(string)
	if targetStr == "" {
		return "", "", 0, fmt.Errorf("environment.ETCD_MEMBER_TARGET is missing or not a string")
	}
	n, convErr := strconv.Atoi(targetStr)
	if convErr != nil {
		return "", "", 0, fmt.Errorf("invalid environment.ETCD_MEMBER_TARGET %q: %w", targetStr, convErr)
	}
	return ip, workDir, n, nil
}

// ---------- RUNNER ----------

type runnerFn func(timeout time.Duration, args ...string) ([]byte, error)

func makeTalosRunner(cli *talosctl.Talosctl, workDir string, logger pulumi.Log) runnerFn {
	return func(timeout time.Duration, args ...string) ([]byte, error) {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		// build "talosctl --talosconfig ... -n <ip> -e <ip> ..."
		full := strings.Fields(cli.BasicCommand)[1:]
		full = append(full, args...)
		logger.Debug(fmt.Sprintf("exec: %s %s", cli.Binary, strings.Join(full, " ")), nil)

		// #nosec G204 — cli.Binary and args are our controlled values
		cmd := exec.CommandContext(ctx, cli.Binary, full...)
		cmd.Dir = workDir

		out, err := cmd.CombinedOutput()
		if err != nil {
			return out, fmt.Errorf("talosctl %v failed: %w: %s", full, err, strings.TrimSpace(string(out)))
		}
		return out, nil
	}
}

func checkEtcdStatus(run runnerFn, timeout time.Duration) error {
	_, err := run(timeout, "etcd", "status")
	return err
}

func listEtcdPeers(run runnerFn, timeout time.Duration) ([]PeerStatus, error) {
	out, err := run(timeout, "etcd", "members")
	if err != nil {
		return nil, err
	}
	peers, perr := parseEtcdPeersFromTable(out)
	if perr != nil {
		return nil, perr
	}
	return peers, nil
}

// peersReady checks count and that no peer is a learner.
func peersReady(peers []PeerStatus, expected int) (bool, string) {
	if len(peers) != expected {
		return false, fmt.Sprintf("expected %d members, got %d", expected, len(peers))
	}
	for _, p := range peers {
		if p.Learner {
			return false, fmt.Sprintf("peer %s is learner", p.ID)
		}
	}
	return true, "ok"
}

/* An example of talosctl etcd members
NODE            ID                 HOSTNAME        PEER URLS                  CLIENT URLS                LEARNER
91.98.138.169   97f365161a13b437   talos-8me-09g   https://10.10.10.2:2380    https://10.10.10.2:2379    false
91.98.138.169   c22a6165837acd5b   talos-apx-beq   https://10.10.10.10:2380   https://10.10.10.10:2379   false
91.98.138.169   d65010d57cf22d49   talos-q2p-yyy   https://10.10.10.5:2380    https://10.10.10.5:2379    false
*/

// sleepWithContext waits for the duration to elapse or the context to be cancelled.
// It returns an error if the context is cancelled before the duration has passed.
func sleepWithContext(ctx context.Context, d time.Duration) error {
	timer := time.NewTimer(d)
	defer timer.Stop()

	select {
	case <-timer.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

/* An example of talosctl etcd members
NODE            ID                 HOSTNAME        PEER URLS                  CLIENT URLS                LEARNER
91.98.138.169   97f365161a13b437   talos-8me-09g   https://10.10.10.2:2380    https://10.10.10.2:2379    false
91.98.138.169   c22a6165837acd5b   talos-apx-beq   https://10.10.10.10:2380   https://10.10.10.10:2379   false
91.98.138.169   d65010d57cf22d49   talos-q2p-yyy   https://10.10.10.5:2380    https://10.10.10.5:2379    false
*/

// parseEtcdPeersFromTable parses `talosctl etcd members` table output and returns peer statuses.
// It skips the header (first non-empty line) and any accidental separator/garbage lines.
func parseEtcdPeersFromTable(out []byte) ([]PeerStatus, error) {
	sc := bufio.NewScanner(bytes.NewReader(out))
	// increase max token size to handle long lines safely (up to ~1MB)
	buf := make([]byte, 0, 64*1024)
	sc.Buffer(buf, 1<<20)

	lines := make([]string, 0, 8)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line != "" {
			lines = append(lines, line)
		}
	}
	if err := sc.Err(); err != nil {
		return nil, fmt.Errorf("scan output: %w", err)
	}
	if len(lines) == 0 {
		return nil, fmt.Errorf("no output from talosctl etcd members")
	}

	// First non-empty line is the header. Everything after is a member row.
	memberRows := lines[1:]
	peers := make([]PeerStatus, 0, len(memberRows))

	for i, row := range memberRows {
		// Expect at least: NODE ID HOSTNAME PEER_URLS CLIENT_URLS LEARNER
		fields := strings.Fields(row)
		if len(fields) < 6 {
			// skip separator/garbage lines
			continue
		}

		id := fields[1]
		learnerStr := fields[len(fields)-1] // last column is LEARNER
		learner, err := strconv.ParseBool(strings.ToLower(learnerStr))
		if err != nil {
			return nil, fmt.Errorf("row %d: invalid LEARNER value %q: %w", i+2, learnerStr, err)
		}

		peers = append(peers, PeerStatus{
			ID:      id,
			Learner: learner,
		})
	}

	if len(peers) == 0 {
		return nil, fmt.Errorf("no peer rows found in talosctl output")
	}
	return peers, nil
}

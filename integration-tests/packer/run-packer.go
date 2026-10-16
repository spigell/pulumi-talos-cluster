// This code is written by AI.
package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

func main() {
	var (
		template    = flag.String("template", "hcloud.pkr.hcl", "Path to Packer template")
		concurrency = flag.Int("concurrency", 4, "Max concurrent packer builds")
		failFast    = flag.Bool("failfast", false, "Cancel remaining builds on first failure")
		timeout     = flag.Duration("timeout", 0, "Optional timeout for each build (e.g. 90m, 2h). 0 = no timeout")
		varList     multiVar
	)
	flag.Var(&varList, "var", "Extra -var key=value (repeatable). Example: -var talos_version=v1.10.3")
	flag.Parse()

	version, err := packerVariable(*template, "talos_version", varList)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	schematic, err := packerVariable(*template, "talos_schematic_id", varList)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if len(schematic) != 64 {
		fmt.Fprintln(os.Stderr, "talos_schematic_id must be 64 characters")
		os.Exit(1)
	}
	token := os.Getenv("HCLOUD_TOKEN")
	if token == "" {
		fmt.Fprintln(os.Stderr, "HCLOUD_TOKEN is required")
		os.Exit(1)
	}
	hcloudClient := &http.Client{Timeout: 30 * time.Second}

	variants := []string{
		"hcloud-amd64",
		"hcloud-arm64",
		"metal-amd64",
		"metal-arm64",
	}

	// Root context (supports Ctrl+C)
	rootCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)

	sem := make(chan struct{}, maxInt(1, *concurrency))
	var wg sync.WaitGroup

	// Track errors and allow fail-fast cancel
	errCh := make(chan error, len(variants))
	ctx := rootCtx
	cancel := func() {}
	if *failFast {
		ctx, cancel = context.WithCancel(rootCtx)
	}

	for _, v := range variants {
		wg.Add(1)
		go func() {
			defer wg.Done()
			select {
			case sem <- struct{}{}:
				defer func() { <-sem }()
			case <-ctx.Done():
				return
			}

			// Per-build context with optional timeout
			runCtx := ctx
			var cancelBuild context.CancelFunc
			if *timeout > 0 {
				runCtx, cancelBuild = context.WithTimeout(ctx, *timeout)
				defer cancelBuild()
			}
			arch := "x86"
			if strings.HasSuffix(v, "arm64") {
				arch = "arm"
			}
			variant := "cloud"
			if strings.HasPrefix(v, "metal-") {
				variant = "metal"
			}
			selector := fmt.Sprintf("os=talos,version=%s,arch=%s,variant=%s,schematic=%s,schematic_suffix=%s", version, arch, variant, schematic[:32], schematic[32:])
			exists, err := snapshotExists(runCtx, hcloudClient, token, selector)
			if err != nil {
				errCh <- fmt.Errorf("[%s] check snapshot: %w", v, err)
				if *failFast {
					cancel()
				}
				return
			}
			if exists {
				fmt.Printf("[%s] matching snapshot exists; skipping build\n", v)
				return
			}

			args := []string{"build", "-var", "target=" + v}
			for _, kv := range varList {
				args = append(args, "-var", kv)
			}
			args = append(args, *template)

			cmd := exec.CommandContext(runCtx, "packer", args...)
			stdout, _ := cmd.StdoutPipe()
			stderr, _ := cmd.StderrPipe()

			if err := cmd.Start(); err != nil {
				errCh <- fmt.Errorf("[%s] start error: %w", v, err)
				if *failFast {
					cancel()
				}
				return
			}

			// Stream logs with prefix
			var logWg sync.WaitGroup
			logWg.Add(2)
			go streamWithPrefix(&logWg, stdout, v)
			go streamWithPrefix(&logWg, stderr, v)

			logWg.Wait()
			if err := cmd.Wait(); err != nil {
				errCh <- fmt.Errorf("[%s] build error: %w", v, err)
				if *failFast {
					cancel()
				}
				return
			}
			fmt.Printf("[%s] ✅ completed\n", v)
		}()
	}

	wg.Wait()
	close(errCh)

	var hadErr bool
	for err := range errCh {
		if err != nil {
			hadErr = true
			fmt.Fprintln(os.Stderr, err)
		}
	}

	if hadErr {
		cancel()
		stop()
		os.Exit(1)
	}
	fmt.Println("All builds finished successfully.")
}

func packerVariable(template, name string, overrides []string) (string, error) {
	for i := len(overrides) - 1; i >= 0; i-- {
		key, value, _ := strings.Cut(overrides[i], "=")
		if key == name {
			return value, nil
		}
	}
	file, diags := hclparse.NewParser().ParseHCLFile(template)
	if diags.HasErrors() {
		return "", fmt.Errorf("parse %s: %s", template, diags.Error())
	}
	body, ok := file.Body.(*hclsyntax.Body)
	if !ok {
		return "", fmt.Errorf("unexpected Packer template body in %s", template)
	}
	for _, block := range body.Blocks {
		if block.Type != "variable" || len(block.Labels) != 1 || block.Labels[0] != name {
			continue
		}
		attribute, ok := block.Body.Attributes["default"]
		if !ok {
			break
		}
		value, diags := attribute.Expr.Value(nil)
		if diags.HasErrors() || !value.IsKnown() || value.IsNull() || value.Type().FriendlyName() != "string" {
			return "", fmt.Errorf("invalid default for %s", name)
		}
		return value.AsString(), nil
	}
	return "", fmt.Errorf("missing default for %s in %s", name, template)
}

func snapshotExists(ctx context.Context, client *http.Client, token, selector string) (bool, error) {
	endpoint := "https://api.hetzner.cloud/v1/images?type=snapshot&per_page=1&label_selector=" + url.QueryEscape(selector)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return false, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("Hetzner image lookup returned HTTP %d", resp.StatusCode)
	}
	var result struct {
		Images *[]struct {
			ID int `json:"id"`
		} `json:"images"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false, fmt.Errorf("decode Hetzner images: %w", err)
	}
	if result.Images == nil {
		return false, fmt.Errorf("Hetzner image lookup omitted images")
	}
	return len(*result.Images) > 0, nil
}

type multiVar []string

func (m *multiVar) String() string { return strings.Join(*m, ",") }
func (m *multiVar) Set(v string) error {
	if !strings.Contains(v, "=") {
		return fmt.Errorf("invalid -var %q, expected key=value", v)
	}
	*m = append(*m, v)
	return nil
}

func streamWithPrefix(wg *sync.WaitGroup, r ioReadCloser, prefix string) {
	defer wg.Done()
	defer r.Close()
	sc := bufio.NewScanner(r)
	buf := make([]byte, 0, 1024*1024)
	sc.Buffer(buf, 1024*1024)
	for sc.Scan() {
		fmt.Printf("[%s] %s\n", prefix, sc.Text())
	}
}

type ioReadCloser interface {
	Read(p []byte) (n int, err error)
	Close() error
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// This code is written by AI.
package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"sync"
	"syscall"
)

func main() {
	var (
		template    = flag.String("template", "hcloud.pkr.hcl", "Path to Packer template")
		concurrency = flag.Int("concurrency", 4, "Max concurrent packer builds")
		failFast    = flag.Bool("failfast", true, "Cancel remaining builds on first failure")
		timeout     = flag.Duration("timeout", 0, "Optional timeout for each build (e.g. 90m, 2h). 0 = no timeout")
		varList     multiVar
	)
	flag.Var(&varList, "var", "Extra -var key=value (repeatable). Example: -var talos_version=v1.10.3")
	flag.Parse()

	variants := []string{
		"hcloud-amd64",
		"hcloud-arm64",
		"metal-amd64",
		"metal-arm64",
	}

	// Root context (supports Ctrl+C)
	rootCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	sem := make(chan struct{}, max(1, *concurrency))
	var wg sync.WaitGroup

	// Track errors and allow fail-fast cancel
	errCh := make(chan error, len(variants))
	ctx := rootCtx
	cancel := func() {}
	if *failFast {
		ctx, cancel = context.WithCancel(rootCtx)
		defer cancel()
	}

	for _, v := range variants {
		v := v // capture
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
		os.Exit(1)
	}
	fmt.Println("All builds finished successfully.")
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

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

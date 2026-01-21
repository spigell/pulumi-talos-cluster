package applier_test

import (
	"os"
	"strconv"

	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
)

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "…"
}

func interpreterFromPV(v resource.PropertyValue) []string {
	if v.IsArray() {
		arr := v.ArrayValue()
		if len(arr) > 0 {
			out := make([]string, 0, len(arr))
			for _, p := range arr {
				if p.IsString() {
					out = append(out, p.StringValue())
				}
			}
			if len(out) > 0 {
				return out
			}
		}
	}
	return []string{"/bin/bash", "-c"}
}

func logLimit() int {
	if v := os.Getenv("TALOSCTL_TEST_LOG_LIMIT"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			return n
		}
	}
	return 200
}

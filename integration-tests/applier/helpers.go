package applier_test

import (
	"os"
	"strconv"

	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
)

func truncate(s string, limit int) string {
	if len(s) <= limit {
		return s
	}
	return s[:limit] + "…"
}

func interpreterFromPV(v resource.PropertyValue) []string {
	if out := flattenStrings(v); len(out) > 0 {
		return out
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

func flattenStrings(v resource.PropertyValue) []string {
	if !v.IsArray() {
		return nil
	}
	arr := v.ArrayValue()
	if len(arr) == 0 {
		return nil
	}
	out := make([]string, 0, len(arr))
	for _, p := range arr {
		if p.IsString() {
			out = append(out, p.StringValue())
		}
	}
	return out
}

package applier

import (
	//"strings"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMergeYAML_MergeSimpleMaps(t *testing.T) {
	yaml1 := `
machine:
  kubelet:
    nodeIP:
      validSubnets: [10.0.0.0/24]
`
	yaml2 := `
cluster:
  proxy:
    image: test-image
`

	result, err := MergeYAML(yaml1, yaml2).Build()
	require.NoError(t, err)
	require.Contains(t, result, "machine:")
	require.Contains(t, result, "cluster:")
	require.Contains(t, result, "proxy:")
	require.NotContains(t, result, "---")
}

func TestMergeYAML_Yaml2IsExtension_ShouldBecomeTail(t *testing.T) {
	yaml1 := `
machine:
  kubelet:
    nodeIP:
      validSubnets: [10.0.0.0/24]
`

	yaml2 := `apiVersion: v1alphav1
kind: ExtensionServiceConfig
metadata:
  name: myext
`

	result, err := MergeYAML(yaml1, yaml2).Build()
	fmt.Println(result)
	require.NoError(t, err)
	require.Contains(t, result, "machine:")
	require.Contains(t, result, "---\napiVersion: v1alphav1")
	require.Contains(t, result, "kind: ExtensionServiceConfig")
}

func TestMergeYAML_Yaml2MissingKind_ShouldError(t *testing.T) {
	yaml1 := `{}`
	yaml2 := `
apiVersion: v1
metadata:
  name: incomplete
`
	_, err := MergeYAML(yaml1, yaml2).Build()
	require.Error(t, err)
	require.Contains(t, err.Error(), "missing kind")
}

func TestMergeYAML_Yaml2MissingApiVersion_ShouldError(t *testing.T) {
	yaml1 := `{}`
	yaml2 := `
kind: ConfigMap
metadata:
  name: incomplete
`
	_, err := MergeYAML(yaml1, yaml2).Build()
	require.Error(t, err)
	require.Contains(t, err.Error(), "missing apiVersion")
}

func TestMergeYAML_Yaml2WithExtraTail(t *testing.T) {
	yaml1 := `
machine:
  time:
    disabled: false
`
	yaml2 := `
debug: true
---
apiVersion: v1
kind: ExtensionServiceConfig
name: cloudflared
`

	result, err := MergeYAML(yaml1, yaml2).Build()
	fmt.Println(result)
	require.NoError(t, err)
	require.Contains(t, result, "machine:")
	require.Contains(t, result, "debug: true")
	require.Contains(t, result, "---\napiVersion: v1")
	require.Contains(t, result, "kind: ExtensionServiceConfig")
}

func TestMergeYAML_Yaml2MultipleDocs_MergeAndTail(t *testing.T) {
	yaml1 := `
machine:
  time:
    disabled: false
`

	yaml2 := `
cluster:
  proxy:
    image: my-proxy
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: test-cm
data:
  key: value
---
debug: true
`

	result, err := MergeYAML(yaml1, yaml2).Build()
	fmt.Println("THE RESULT:")
	fmt.Println(result)
	require.NoError(t, err)

	// Head should include merged config from first and third doc
	require.Contains(t, result, "machine:")
	require.Contains(t, result, "cluster:")
	require.Contains(t, result, "debug: true")

	// Tail should include ConfigMap only
	docCount := strings.Count(result, "---")
	require.Equal(t, 1, docCount, "only one tail document (ConfigMap) expected")

	require.Contains(t, result, "kind: ConfigMap")
	require.Contains(t, result, "name: test-cm")
}

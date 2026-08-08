package applier

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBuildSecureApplyConfig_DoesNotDuplicateGeneratedPatchArrays(t *testing.T) {
	machineConfig := `
machine:
  install:
    image: old-installer
  kubelet:
    image: kubelet:v1
    nodeIP:
      validSubnets:
        - 192.168.1.0/24
  registries:
    mirrors:
      registry.local-image-builder.svc.cluster.local:5000:
        endpoints:
          - http://10.111.64.155:5000
        overridePath: true
cluster:
  apiServer:
    image: api-server:v1
  controllerManager:
    image: controller-manager:v1
  scheduler:
    image: scheduler:v1
  proxy:
    image: proxy:v1
`

	result, err := buildSecureApplyConfig(machineConfig, "new-installer", &K8SImages{
		Kubelet:           "kubelet:v1",
		APIServer:         "api-server:v1",
		ControllerManager: "controller-manager:v1",
		Scheduler:         "scheduler:v1",
		KubeProxy:         "proxy:v1",
	})

	require.NoError(t, err)
	require.Contains(t, result, "image: new-installer")
	require.Equal(t, 1, strings.Count(result, "192.168.1.0/24"))
	require.Equal(t, 1, strings.Count(result, "http://10.111.64.155:5000"))
}

package applier

import (
	"testing"

	"github.com/siderolabs/talos/pkg/machinery/config/types/v1alpha1"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestNewK8SImages(t *testing.T) {
	const configYAML = `
machine:
  type: controlplane
  kubelet:
    image: kubelet-image
cluster:
  apiServer:
    image: apiserver-image
  proxy:
    image: proxy-image
  scheduler:
    image: scheduler-image
  controllerManager:
    image: controller-manager-image
`
	var config v1alpha1.Config
	require.NoError(t, yaml.Unmarshal([]byte(configYAML), &config))
	require.Equal(t, &K8SImages{
		Kubelet:           "kubelet-image",
		APIServer:         "apiserver-image",
		KubeProxy:         "proxy-image",
		Scheduler:         "scheduler-image",
		ControllerManager: "controller-manager-image",
	}, NewK8SImages(&config))

	config.MachineConfig.MachineType = "worker"
	require.Equal(t, &K8SImages{Kubelet: "kubelet-image"}, NewK8SImages(&config))
}

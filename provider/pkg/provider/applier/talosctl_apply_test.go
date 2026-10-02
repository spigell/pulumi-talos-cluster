package applier

import (
	"strings"
	"testing"

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
	images, err := NewK8SImages(configYAML)
	require.NoError(t, err)
	require.Equal(t, &K8SImages{
		Kubelet:           "kubelet-image",
		APIServer:         "apiserver-image",
		KubeProxy:         "proxy-image",
		Scheduler:         "scheduler-image",
		ControllerManager: "controller-manager-image",
	}, images)

	workerImages, err := NewK8SImages(strings.Replace(configYAML, "type: controlplane", "type: worker", 1))
	require.NoError(t, err)
	require.Equal(t, &K8SImages{Kubelet: "kubelet-image"}, workerImages)
}

func TestNewK8SImagesPreservesAbsentFields(t *testing.T) {
	const current = `
machine:
  type: controlplane
  kubelet:
    image: current-kubelet
cluster:
  apiServer: {}
  proxy:
    image: current-proxy
`
	images, err := NewK8SImages(current)
	require.NoError(t, err)
	require.Equal(t, &K8SImages{Kubelet: "current-kubelet", KubeProxy: "current-proxy"}, images)

	const proposed = `
machine:
  kubelet:
    image: proposed-kubelet
cluster:
  apiServer:
    image: proposed-api
  proxy:
    image: proposed-proxy
  scheduler:
    image: proposed-scheduler
  controllerManager:
    image: proposed-controller
`
	result, err := MergeYAML(proposed, "{}").WithGuard(GuardUnmodifyK8sImages(images)).Build()
	require.NoError(t, err)
	var merged map[string]any
	require.NoError(t, yaml.Unmarshal([]byte(result), &merged))
	machine := merged["machine"].(map[string]any)
	cluster := merged["cluster"].(map[string]any)
	require.Equal(t, "current-kubelet", machine["kubelet"].(map[string]any)["image"])
	require.Equal(t, "current-proxy", cluster["proxy"].(map[string]any)["image"])
	for _, component := range []string{"apiServer", "scheduler", "controllerManager"} {
		require.NotContains(t, cluster[component].(map[string]any), "image")
	}

	emptyImages, err := NewK8SImages("machine:\n  type: controlplane\ncluster: {}\n")
	require.NoError(t, err)
	result, err = MergeYAML(proposed, "{}").WithGuard(GuardUnmodifyK8sImages(emptyImages)).Build()
	require.NoError(t, err)
	require.NotContains(t, result, "image:")
}

package provider

import "testing"

func TestClusterApplyContractConstants(t *testing.T) {
	expected := map[string]string{
		"ClusterResourceOutputsMachines":                                "machines",
		"ClusterResourceOutputsGeneratedConfigurations":                 "generatedConfigurations",
		"ClusterResourceOutputsControlplaneMachineConfigurations":       "controlplaneMachineConfigurations",
		"ClusterResourceOutputsWorkerMachineConfigurations":             "workerMachineConfigurations",
		"ClusterResourceOutputsInitMachineConfiguration":                "initMachineConfiguration",
		"ClusterResourceOutputsClientConfiguration":                     "clientConfiguration",
		"ClusterResourceOutputsClientConfigurationCAKey":                "caCertificate",
		"ClusterResourceOutputsClientConfigurationClientKey":            "clientKey",
		"ClusterResourceOutputsClientConfigurationClientCertificateKey": "clientCertificate",
	}

	actual := map[string]string{
		"ClusterResourceOutputsMachines":                                ClusterResourceOutputsMachines,
		"ClusterResourceOutputsGeneratedConfigurations":                 ClusterResourceOutputsGeneratedConfigurations,
		"ClusterResourceOutputsControlplaneMachineConfigurations":       ClusterResourceOutputsControlplaneMachineConfigurations,
		"ClusterResourceOutputsWorkerMachineConfigurations":             ClusterResourceOutputsWorkerMachineConfigurations,
		"ClusterResourceOutputsInitMachineConfiguration":                ClusterResourceOutputsInitMachineConfiguration,
		"ClusterResourceOutputsClientConfiguration":                     ClusterResourceOutputsClientConfiguration,
		"ClusterResourceOutputsClientConfigurationCAKey":                ClusterResourceOutputsClientConfigurationCAKey,
		"ClusterResourceOutputsClientConfigurationClientKey":            ClusterResourceOutputsClientConfigurationClientKey,
		"ClusterResourceOutputsClientConfigurationClientCertificateKey": ClusterResourceOutputsClientConfigurationClientCertificateKey,
	}

	for name, expectedValue := range expected {
		if actualValue := actual[name]; actualValue != expectedValue {
			t.Fatalf("contract constant %s mismatch: expected %q, got %q", name, expectedValue, actualValue)
		}
	}
}

func TestClusterApplyResourceTypeNames(t *testing.T) {
	if got := ClusterType(); got != "talos-cluster:index:Cluster" {
		t.Fatalf("cluster resource type changed: %q", got)
	}

	if got := ApplyType(); got != "talos-cluster:index:Apply" {
		t.Fatalf("apply resource type changed: %q", got)
	}
}

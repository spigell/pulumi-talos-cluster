package provider

import (
	"fmt"
	"testing"

	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

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

func TestClusterApplyClientConfigurationContract(t *testing.T) {
	err := pulumi.RunErr(func(ctx *pulumi.Context) error {
		input := pulumi.StringMap{
			ClusterResourceOutputsClientConfigurationCAKey:                pulumi.String("ca"),
			ClusterResourceOutputsClientConfigurationClientKey:            pulumi.String("client-key"),
			ClusterResourceOutputsClientConfigurationClientCertificateKey: pulumi.String("client-cert"),
		}.ToStringMapOutput()

		conf := buildClientConfigurationFromMap(input)

		var assertErr error
		pulumi.All(conf.CaCertificate, conf.ClientKey, conf.ClientCertificate).ApplyT(func(vals []any) error {
			if vals[0].(string) != "ca" {
				assertErr = fmt.Errorf("caCertificate mismatch: got %q", vals[0].(string))
			}
			if vals[1].(string) != "client-key" {
				assertErr = fmt.Errorf("clientKey mismatch: got %q", vals[1].(string))
			}
			if vals[2].(string) != "client-cert" {
				assertErr = fmt.Errorf("clientCertificate mismatch: got %q", vals[2].(string))
			}
			return nil
		})

		return assertErr
	}, pulumi.WithMocks("contract-project", "dev", &contractMocks{}))
	if err != nil {
		t.Fatalf("contract between Cluster and Apply client configuration changed: %v", err)
	}
}

type contractMocks struct{}

func (contractMocks) NewResource(args pulumi.MockResourceArgs) (string, resource.PropertyMap, error) {
	return fmt.Sprintf("%s-id", args.Name), args.Inputs, nil
}

func (contractMocks) Call(args pulumi.MockCallArgs) (resource.PropertyMap, error) {
	return resource.PropertyMap{}, nil
}

package cluster

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/santhosh-tekuri/jsonschema/v5"
	"gopkg.in/yaml.v3"
)

func parseToMap(data []byte) (map[string]any, error) {
	var raw map[string]any
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, err
	}
	return raw, nil
}

func validateCluster(raw map[string]any) error {
	schema, err := loadSchema()
	if err != nil {
		return err
	}
	if err := schema.Validate(raw); err != nil {
		return fmt.Errorf("invalid cluster spec: %w", err)
	}

	machinesVal := raw["machines"].([]any) // safe after schema validation

	usePrivateNetwork := raw["usePrivateNetwork"].(bool) // safe

	validationCidr := ""
	if usePrivateNetwork {
		privateNetwork := ""
		if v, ok := raw["privateNetwork"].(string); ok {
			privateNetwork = strings.TrimSpace(v)
		}
		privateSubnetwork := ""
		if v, ok := raw["privateSubnetwork"].(string); ok {
			privateSubnetwork = strings.TrimSpace(v)
		}

		if privateNetwork == "" || privateSubnetwork == "" {
			return fmt.Errorf("when 'usePrivateNetwork' is true, both 'privateNetwork' and 'privateSubnetwork' are required")
		}

		validationCidr = privateSubnetwork
	}

	for _, m := range machinesVal {
		mv, _ := m.(map[string]any)
		if err := validateMachine(mv, validationCidr); err != nil {
			return err
		}
	}

	return nil
}

func validateMachine(raw map[string]any, validationCidr string) error {
	id := raw["id"].(string) // required by schema
	privateIP := ""
	if v, ok := raw["privateIP"].(string); ok {
		privateIP = strings.TrimSpace(v)
	}

	if validationCidr != "" {
		if privateIP == "" {
			return fmt.Errorf("invalid cluster spec: machine %q must define privateIP", id)
		}
		if err := assertIPInNetwork(privateIP, validationCidr); err != nil {
			return fmt.Errorf("invalid cluster spec: machine '%s' privateIP '%s' must be inside '%s'", id, privateIP, validationCidr)
		}
	}

	return nil
}

func assertIPInNetwork(ipStr, cidr string) error {
	ip := net.ParseIP(ipStr)
	_, network, err := net.ParseCIDR(cidr)
	if err != nil {
		return fmt.Errorf("invalid cidr: %w", err)
	}
	if ip == nil || !network.Contains(ip) {
		return fmt.Errorf("ip not in network")
	}
	return nil
}

var (
	schemaOnce sync.Once
	compiled   *jsonschema.Schema
	schemaErr  error
)

func loadSchema() (*jsonschema.Schema, error) {
	schemaOnce.Do(func() {
		_, file, _, ok := runtime.Caller(0)
		if !ok {
			schemaErr = fmt.Errorf("cannot resolve schema path")
			return
		}
		schemaPath := filepath.Join(filepath.Dir(file), "..", "schema.json")

		f, err := os.Open(schemaPath)
		if err != nil {
			schemaErr = fmt.Errorf("load schema: %w", err)
			return
		}
		defer f.Close()

		compiler := jsonschema.NewCompiler()
		compiler.Draft = jsonschema.Draft7
		if err := compiler.AddResource("schema.json", f); err != nil {
			schemaErr = fmt.Errorf("load schema: %w", err)
			return
		}
		compiled, schemaErr = compiler.Compile("schema.json")
	})
	return compiled, schemaErr
}

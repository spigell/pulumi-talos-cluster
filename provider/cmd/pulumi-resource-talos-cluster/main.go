//go:generate go run ./generate.go

package main

import (
	"github.com/spigell/pulumi-talos-cluster/provider/pkg/provider"
	"github.com/spigell/pulumi-talos-cluster/provider/pkg/version"
)

func main() {
	provider.Serve(version.Version, pulumiSchema)
}

package cloud

import "github.com/pulumi/pulumi/sdk/v3/go/pulumi"

// Provider defines an interface for cloud providers that can provision
// networks and servers for the integration tests.
type Provider interface {
	// Servers returns the list of servers that will be created by the provider.
	Servers() []Server
	// Up provisions the infrastructure in the cloud provider.
	Up() (*Deployed, error)
}

// Server represents a machine instance created by a cloud provider.
type Server interface {
	// ID returns the identifier for the server.
	ID() string
	// IP returns the public IPv4 address assigned to the server.
	IP() pulumi.StringOutput
	// WithUserdata sets the user data for the server and returns the updated
	// server instance.
	WithUserdata(userdata pulumi.StringOutput) Server
}

// Network abstracts a private network created by the cloud provider.
type Network interface {
	// ID returns the identifier for the network resource.
	ID() pulumi.StringOutput
}

// Deployed represents resources created by a Provider.
type Deployed struct {
	Servers []Server
	Deps    []pulumi.Resource
}

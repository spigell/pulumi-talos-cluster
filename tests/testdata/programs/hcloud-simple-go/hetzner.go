package main

import (
	"fmt"
	"strconv"

	"github.com/pulumi/pulumi-hcloud/sdk/go/hcloud"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type DeployedServer struct {
	Name      string
	IP        pulumi.StringOutput
	PrivateIP pulumi.StringOutput
}

func NewHetzner(ctx *pulumi.Context, cluster *Cluster) ([]*DeployedServer, error) {
	network, err := hcloud.NewNetwork(ctx, "private-network", &hcloud.NetworkArgs{
		Name:    pulumi.Sprintf("private-network-%s", cluster.Name),
		IpRange: pulumi.String(cluster.PrivateNetwork), // Define the CIDR block for the network
	})
	if err != nil {
		return nil, fmt.Errorf("error creating network: %w", err)
	}

	convertedNetID := network.ID().ApplyT(func(id string) (int, error) {
		return strconv.Atoi(id)
	}).(pulumi.IntOutput)

	// Add a subnet to the private network
	_, err = hcloud.NewNetworkSubnet(ctx, "private-subnet", &hcloud.NetworkSubnetArgs{
		NetworkId:   convertedNetID,
		Type:        pulumi.String("server"),
		NetworkZone: pulumi.String("eu-central"),              // Adjust based on your preferred region
		IpRange:     pulumi.String(cluster.PrivateSubnetwork), // Define the CIDR block for the subnet
	})
	if err != nil {
		return nil, fmt.Errorf("error creating subnet: %w", err)
	}


	selector := "os=talos,testing=true"
	image, err := hcloud.GetImage(ctx, &hcloud.GetImageArgs{
		WithSelector: pulumi.StringRef(selector),
		MostRecent:   pulumi.BoolRef(true),
		WithArchitecture: pulumi.StringRef("x86"),
	})

	if err != nil {
		return nil, fmt.Errorf("can't find an image")
	}

	deployed := make([]*DeployedServer, 0)

	for _, machine := range cluster.Machines {
		// Define the server
		server, err := hcloud.NewServer(ctx, machine.Name, &hcloud.ServerArgs{
			Name:       pulumi.String(machine.Name),
			ServerType: pulumi.String(machine.ServerType),
			// Image:      pulumi.Sprintf(fmt.Sprintf("%d", image.Id)), // OS image
			Image:      pulumi.Sprintf("%d", image.Id), // OS image
			Location:   pulumi.String("nbg1"),                   // Choose the Hetzner location
			Networks: &hcloud.ServerNetworkTypeArray{
				hcloud.ServerNetworkTypeArgs{
					NetworkId: convertedNetID,
					Ip:        pulumi.String(machine.PrivateIP),
				},
			},
		})
		if err != nil {
			return nil, fmt.Errorf("error creating server: %w", err)
		}

		deployed = append(deployed, &DeployedServer{
			Name: machine.Name,
			IP:   server.Ipv4Address,
		})
	}

	return deployed, nil
}

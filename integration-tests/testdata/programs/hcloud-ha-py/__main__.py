from pathlib import Path

import pulumi

from hetzner import hetzner
from cluster.python.spec import load  # provided by installed integration-tests package
from pulumi_talos_cluster import Apply, Cluster as TalosCluster, MachineTypes

base_dir = Path(__file__).resolve().parent
cluster_path = base_dir / "cluster.yaml"
cluster = load(str(cluster_path))

servers = hetzner(cluster)
machines_by_id = {srv["id"]: srv for srv in servers}

cluster_machines = []
for m in cluster.machines:
    srv = machines_by_id.get(m.id)
    if srv is None:
        raise ValueError(f"server for machine {m.id} not found")

    cluster_machines.append(
        {
            "machine_id": m.id,
            "node_ip": srv["ip"],
            "machine_type": MachineTypes(m.type),
            "config_patches": m.configPatches or None,
            "talos_image": m.talosImage or None,
        }
    )

cluster_endpoint = pulumi.Output.concat("https://", cluster_machines[0]["node_ip"], ":6443")

talos_cluster = TalosCluster(
    cluster.name,
    cluster_endpoint=cluster_endpoint,
    cluster_machines=cluster_machines,
    cluster_name=cluster.name,
    kubernetes_version=cluster.kubernetesVersion,
)

apply = Apply(
    cluster.name,
    client_configuration=talos_cluster.client_configuration,
    apply_machines=talos_cluster.machines,
    skip_init_apply=cluster.skipInitApply,
)

pulumi.export("clusterName", cluster.name)
pulumi.export("machineIds", list(machines_by_id.keys()))
pulumi.export("clusterMachineConfigs", talos_cluster.generated_configurations)
pulumi.export("kubeconfig", apply.credentials.kubeconfig)
pulumi.export("talosconfig", apply.credentials.talosconfig)

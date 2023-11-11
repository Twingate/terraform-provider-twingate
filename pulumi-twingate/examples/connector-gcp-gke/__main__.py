import pulumi
from pulumi_gcp import compute, container
import pulumi_twingate as tg
import pulumi_kubernetes as kubernetes
import os
from pulumi_kubernetes.helm.v3 import Release, ReleaseArgs, RepositoryOptsArgs

config = pulumi.Config()
data = config.require_object("data")
gcp_config = pulumi.Config("gcp")
twingate_config = pulumi.Config("twingate")

try:
    tg_account = twingate_config.get("network")
    if tg_account is None:
        tg_account = os.getenv('TWINGATE_NETWORK')
except:
    tg_account = os.getenv('TWINGATE_NETWORK')

# Create an VPC
vpc = compute.Network(
    data.get("vpc_name"),
    name=data.get("vpc_name"),
    auto_create_subnetworks=False,
)

# Create a Subnet
subnet = compute.Subnetwork(data.get("subnet_name"),
                            name=data.get("subnet_name"),
                            ip_cidr_range=data.get("subnet_cidr"),
                            network=vpc.id,
                            )

# Create a Cloud Router
router = compute.Router(data.get("router_name"),
                        name=data.get("router_name"),
                        network=vpc.id
                        )
# Create a Cloud NAT
nat = compute.RouterNat(data.get("nat_name"),
                        name=data.get("nat_name"),
                        router=router.name,
                        nat_ip_allocate_option="AUTO_ONLY",
                        source_subnetwork_ip_ranges_to_nat="ALL_SUBNETWORKS_ALL_IP_RANGES",
                        log_config=compute.RouterNatLogConfigArgs(
                            enable=True,
                            filter="ERRORS_ONLY",
                        ))

# Create a GKE cluster
cluster = container.Cluster(data.get("cluster_name"),
                            name=data.get("cluster_name"),
                            remove_default_node_pool=True,
                            initial_node_count=1,
                            min_master_version=data.get("master_version"),
                            network=vpc.id,
                            subnetwork=subnet.id,
                            private_cluster_config=container.ClusterPrivateClusterConfigArgs(
                                enable_private_nodes=True,
                                master_ipv4_cidr_block="172.16.0.0/28",
                            ),
                            ip_allocation_policy=container.ClusterIpAllocationPolicyArgs()
                            )

# Create a GKE nodepool
node_pool = container.NodePool(data.get("node_pool_name"),
                               name=data.get("node_pool_name"),
                               cluster=cluster.name,
                               node_count=data.get("node_count"),
                               node_config=container.NodePoolNodeConfigArgs(
                                   machine_type=data.get("node_machine_type"),
                               ),
                               opts=pulumi.ResourceOptions(depends_on=[cluster],
                                                           custom_timeouts=pulumi.CustomTimeouts(create='30m'))
                               )

# Construct K8S configuration
cluster_info = pulumi.Output.all(cluster.name, cluster.endpoint, cluster.master_auth)
cluster_config = cluster_info.apply(
    lambda info: """apiVersion: v1
clusters:
- cluster:
    certificate-authority-data: {0}
    server: https://{1}
  name: {2}
contexts:
- context:
    cluster: {2}
    user: {2}
  name: {2}
current-context: {2}
kind: Config
preferences: {{}}
users:
- name: {2}
  user:
    exec:
      apiVersion: client.authentication.k8s.io/v1beta1
      command: gke-gcloud-auth-plugin
      installHint: Install gke-gcloud-auth-plugin for use with kubectl by following
        https://cloud.google.com/blog/products/containers-kubernetes/kubectl-auth-changes-in-gke
      provideClusterInfo: true
""".format(info[2]['cluster_ca_certificate'],
           info[1], '{0}_{1}_{2}'.format(gcp_config.get("project"), gcp_config.get("zone"), info[0])))

# Create K8S provider
cluster_provider = kubernetes.Provider('gke_k8s',
                                       kubeconfig=cluster_config,
                                       opts=pulumi.ResourceOptions(depends_on=[cluster],
                                                                   custom_timeouts=pulumi.CustomTimeouts(create='30m')))

# Create a Twingate Remote Network
remote_network = tg.TwingateRemoteNetwork(data.get("tg_remote_network"), name=data.get("tg_remote_network"))

connectors = data.get("connectors")

# Create a Pod For Each Connector
for i in range(1, connectors + 1):
    connector = tg.TwingateConnector(f"connector_{i}", name="", remote_network_id=remote_network.id)
    connector_token = tg.TwingateConnectorTokens(f"demo_token_{i}", connector_id=connector.id)

    # Deploying Helm Chart to the Kubernetes Cluster
    chart = Release(
        f"twingate-connector-{i}",
        ReleaseArgs(
            chart="connector",
            name=pulumi.Output.all(connector.name).apply(
                lambda v: f"tg-{v[0]}"),
            namespace=data.get("namespace"),
            repository_opts=RepositoryOptsArgs(
                repo="https://twingate.github.io/helm-charts",
            ),
            values={
                "connector": {
                    "network": tg_account,
                    "accessToken": connector_token.access_token,
                    "refreshToken": connector_token.refresh_token
                },
                "additionalLabels": {
                    "app": "twingate-connector"
                },
                "affinity": {
                    "podAntiAffinity": {
                        "preferredDuringSchedulingIgnoredDuringExecution": [
                            {
                                "weight": 1,
                                "podAffinityTerm": {
                                    "labelSelector": {
                                        "matchExpressions": [
                                            {"key": "app", "operator": "In", "values": ["twingate-connector"]}
                                        ]
                                    },
                                    "topologyKey": "kubernetes.io/hostname"
                                }
                            }
                        ]
                    }
                }
            },
            timeout=1800
        ),
        opts=pulumi.ResourceOptions(depends_on=[node_pool, cluster, nat],
                                    custom_timeouts=pulumi.CustomTimeouts(create='30m'),
                                    provider=cluster_provider),
    )

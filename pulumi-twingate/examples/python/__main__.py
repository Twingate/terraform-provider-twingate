import pulumi
import pulumi_twingate as tg

# Create a Twingate remote network
remote_network = tg.TwingateRemoteNetwork("test_network", name="Office")

# Create a Twingate service account
service_account = tg.TwingateServiceAccount("ci_cd_account", name="CI CD Service")

# Create a Twingate service key
service_account_key = tg.TwingateServiceAccountKey("ci_cd_key", name="CI CD Key", service_account_id=service_account.id)

# To see service_account_key, execute command `pulumi stack output --show-secrets`
pulumi.export("service_account_key", service_account_key.token)


# Get group id by name
def get_group_id(group_name):
    group = tg.get_twingate_groups_output(name=group_name).groups[0]
    return group.id


# Create a Twingate Resource and configure resource permission
twingate_resource = tg.TwingateResource("twingate_home_page",
                                        name="Twingate Home Page",
                                        address="www.twingate.com",
                                        remote_network_id=remote_network.id,
                                        access={"group_ids": [get_group_id("Everyone")],
                                                "service_account_ids": [service_account.id]}
                                        )
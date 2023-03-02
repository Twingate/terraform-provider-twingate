---
title: Twingate
meta_desc: Provides an overview of the Twingate Provider for Pulumi.
layout: overview
---

The Twingate provider for Pulumi can be used to provision any of the cloud resources available in [Twingate](https://www.twingate.com/).
The Twingate provider must be configured with credentials to deploy and update resources in Twingate.

## Example

{{< chooser language "typescript,python,csharp" >}}
{{% choosable language typescript %}}

```typescript
import * as tg from "@twingate-labs/pulumi-twingate"
import * as pulumi from "@pulumi/pulumi"

const remoteNetwork = new tg.TwingateRemoteNetwork("test-network", {name: "Pulumi Test Network"})
const serviceAccount = new tg.TwingateServiceAccount("ci_cd_account", {name: "CI CD Service"})
const serviceAccountKey = new tg.TwingateServiceAccountKey("ci_cd_key", {name: "CI CD Key", serviceAccountId: serviceAccount.id})

// To see serviceAccountKeyOut, execute command `pulumi stack output --show-secrets`
export const serviceAccountKeyOut = pulumi.interpolate`${serviceAccountKey.token}`;

// get group id by name
function getGroupId(groupName: string){
    const groups:any = tg.getTwingateGroupsOutput({name: groupName})?.groups ?? []
    return groups[0].id
}

new tg.TwingateResource("test_resource", {
    name: "Twingate Home Page",
    address: "www.twingate.com",
    remoteNetworkId: remoteNetwork.id,
    access: {
        groupIds: [getGroupId("Everyone")],
        serviceAccountIds: [serviceAccount.id]
    }
})
```

{{% /choosable %}}
{{% choosable language python %}}

```python
import pulumi
import pulumi_twingate as tg

remote_network = tg.TwingateRemoteNetwork("test_network", name="Pulumi Test Network")
service_account = tg.TwingateServiceAccount("ci_cd_account", name="CI CD Service")
service_account_key = tg.TwingateServiceAccountKey("ci_cd_key", name="CI CD Key", service_account_id=service_account.id)

# To see service_account_key, execute command `pulumi stack output --show-secrets`
pulumi.export("service_account_key", service_account_key.token)


# Get group id by name
def get_group_id(group_name):
    group = tg.get_twingate_groups_output(name=group_name).groups[0]
    return group.id


twingate_resource = tg.TwingateResource("test_resource",
                                        name="Twingate Home Page",
                                        address="www.twingate.com",
                                        remote_network_id=remote_network.id,
                                        access={"group_ids": [get_group_id("Everyone")],
                                                "service_account_ids": [service_account.id]}
                                        )
```

{{% /choosable %}}

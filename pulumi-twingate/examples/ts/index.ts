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
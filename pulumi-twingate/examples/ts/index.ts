import * as twingate from "@twingate-labs/pulumi-twingate"

const remoteNetwork = new twingate.TwingateRemoteNetwork("test-network", {
    name: "Test Network"
})

new twingate.TwingateResource("test-resource", {
    name: "Pulumi Website",
    address: "www.pulumi.com",
    remoteNetworkId: remoteNetwork.id,
    groupIds: ["R3JvdXA6MzA2MDk="]
})

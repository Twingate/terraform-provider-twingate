# coding=utf-8
# *** WARNING: this file was generated by the Pulumi Terraform Bridge (tfgen) Tool. ***
# *** Do not edit by hand unless you're certain you know what you are doing! ***

import copy
import warnings
import pulumi
import pulumi.runtime
from typing import Any, Mapping, Optional, Sequence, Union, overload
from . import _utilities

__all__ = [
    'TwingateResourceProtocolsArgs',
    'TwingateResourceProtocolsTcpArgs',
    'TwingateResourceProtocolsUdpArgs',
    'GetTwingateConnectorsConnectorArgs',
    'GetTwingateGroupsGroupArgs',
    'GetTwingateResourceProtocolArgs',
    'GetTwingateResourceProtocolTcpArgs',
    'GetTwingateResourceProtocolUdpArgs',
    'GetTwingateResourcesResourceArgs',
    'GetTwingateResourcesResourceProtocolArgs',
    'GetTwingateResourcesResourceProtocolTcpArgs',
    'GetTwingateResourcesResourceProtocolUdpArgs',
    'GetTwingateUsersUserArgs',
]

@pulumi.input_type
class TwingateResourceProtocolsArgs:
    def __init__(__self__, *,
                 tcp: pulumi.Input['TwingateResourceProtocolsTcpArgs'],
                 udp: pulumi.Input['TwingateResourceProtocolsUdpArgs'],
                 allow_icmp: Optional[pulumi.Input[bool]] = None):
        pulumi.set(__self__, "tcp", tcp)
        pulumi.set(__self__, "udp", udp)
        if allow_icmp is not None:
            pulumi.set(__self__, "allow_icmp", allow_icmp)

    @property
    @pulumi.getter
    def tcp(self) -> pulumi.Input['TwingateResourceProtocolsTcpArgs']:
        return pulumi.get(self, "tcp")

    @tcp.setter
    def tcp(self, value: pulumi.Input['TwingateResourceProtocolsTcpArgs']):
        pulumi.set(self, "tcp", value)

    @property
    @pulumi.getter
    def udp(self) -> pulumi.Input['TwingateResourceProtocolsUdpArgs']:
        return pulumi.get(self, "udp")

    @udp.setter
    def udp(self, value: pulumi.Input['TwingateResourceProtocolsUdpArgs']):
        pulumi.set(self, "udp", value)

    @property
    @pulumi.getter(name="allowIcmp")
    def allow_icmp(self) -> Optional[pulumi.Input[bool]]:
        return pulumi.get(self, "allow_icmp")

    @allow_icmp.setter
    def allow_icmp(self, value: Optional[pulumi.Input[bool]]):
        pulumi.set(self, "allow_icmp", value)


@pulumi.input_type
class TwingateResourceProtocolsTcpArgs:
    def __init__(__self__, *,
                 policy: pulumi.Input[str],
                 ports: Optional[pulumi.Input[Sequence[pulumi.Input[str]]]] = None):
        pulumi.set(__self__, "policy", policy)
        if ports is not None:
            pulumi.set(__self__, "ports", ports)

    @property
    @pulumi.getter
    def policy(self) -> pulumi.Input[str]:
        return pulumi.get(self, "policy")

    @policy.setter
    def policy(self, value: pulumi.Input[str]):
        pulumi.set(self, "policy", value)

    @property
    @pulumi.getter
    def ports(self) -> Optional[pulumi.Input[Sequence[pulumi.Input[str]]]]:
        return pulumi.get(self, "ports")

    @ports.setter
    def ports(self, value: Optional[pulumi.Input[Sequence[pulumi.Input[str]]]]):
        pulumi.set(self, "ports", value)


@pulumi.input_type
class TwingateResourceProtocolsUdpArgs:
    def __init__(__self__, *,
                 policy: pulumi.Input[str],
                 ports: Optional[pulumi.Input[Sequence[pulumi.Input[str]]]] = None):
        pulumi.set(__self__, "policy", policy)
        if ports is not None:
            pulumi.set(__self__, "ports", ports)

    @property
    @pulumi.getter
    def policy(self) -> pulumi.Input[str]:
        return pulumi.get(self, "policy")

    @policy.setter
    def policy(self, value: pulumi.Input[str]):
        pulumi.set(self, "policy", value)

    @property
    @pulumi.getter
    def ports(self) -> Optional[pulumi.Input[Sequence[pulumi.Input[str]]]]:
        return pulumi.get(self, "ports")

    @ports.setter
    def ports(self, value: Optional[pulumi.Input[Sequence[pulumi.Input[str]]]]):
        pulumi.set(self, "ports", value)


@pulumi.input_type
class GetTwingateConnectorsConnectorArgs:
    def __init__(__self__, *,
                 id: str,
                 name: str,
                 remote_network_id: str):
        pulumi.set(__self__, "id", id)
        pulumi.set(__self__, "name", name)
        pulumi.set(__self__, "remote_network_id", remote_network_id)

    @property
    @pulumi.getter
    def id(self) -> str:
        return pulumi.get(self, "id")

    @id.setter
    def id(self, value: str):
        pulumi.set(self, "id", value)

    @property
    @pulumi.getter
    def name(self) -> str:
        return pulumi.get(self, "name")

    @name.setter
    def name(self, value: str):
        pulumi.set(self, "name", value)

    @property
    @pulumi.getter(name="remoteNetworkId")
    def remote_network_id(self) -> str:
        return pulumi.get(self, "remote_network_id")

    @remote_network_id.setter
    def remote_network_id(self, value: str):
        pulumi.set(self, "remote_network_id", value)


@pulumi.input_type
class GetTwingateGroupsGroupArgs:
    def __init__(__self__, *,
                 id: str,
                 is_active: bool,
                 name: str,
                 type: str):
        pulumi.set(__self__, "id", id)
        pulumi.set(__self__, "is_active", is_active)
        pulumi.set(__self__, "name", name)
        pulumi.set(__self__, "type", type)

    @property
    @pulumi.getter
    def id(self) -> str:
        return pulumi.get(self, "id")

    @id.setter
    def id(self, value: str):
        pulumi.set(self, "id", value)

    @property
    @pulumi.getter(name="isActive")
    def is_active(self) -> bool:
        return pulumi.get(self, "is_active")

    @is_active.setter
    def is_active(self, value: bool):
        pulumi.set(self, "is_active", value)

    @property
    @pulumi.getter
    def name(self) -> str:
        return pulumi.get(self, "name")

    @name.setter
    def name(self, value: str):
        pulumi.set(self, "name", value)

    @property
    @pulumi.getter
    def type(self) -> str:
        return pulumi.get(self, "type")

    @type.setter
    def type(self, value: str):
        pulumi.set(self, "type", value)


@pulumi.input_type
class GetTwingateResourceProtocolArgs:
    def __init__(__self__, *,
                 allow_icmp: bool,
                 tcps: Optional[Sequence['GetTwingateResourceProtocolTcpArgs']] = None,
                 udps: Optional[Sequence['GetTwingateResourceProtocolUdpArgs']] = None):
        pulumi.set(__self__, "allow_icmp", allow_icmp)
        if tcps is not None:
            pulumi.set(__self__, "tcps", tcps)
        if udps is not None:
            pulumi.set(__self__, "udps", udps)

    @property
    @pulumi.getter(name="allowIcmp")
    def allow_icmp(self) -> bool:
        return pulumi.get(self, "allow_icmp")

    @allow_icmp.setter
    def allow_icmp(self, value: bool):
        pulumi.set(self, "allow_icmp", value)

    @property
    @pulumi.getter
    def tcps(self) -> Optional[Sequence['GetTwingateResourceProtocolTcpArgs']]:
        return pulumi.get(self, "tcps")

    @tcps.setter
    def tcps(self, value: Optional[Sequence['GetTwingateResourceProtocolTcpArgs']]):
        pulumi.set(self, "tcps", value)

    @property
    @pulumi.getter
    def udps(self) -> Optional[Sequence['GetTwingateResourceProtocolUdpArgs']]:
        return pulumi.get(self, "udps")

    @udps.setter
    def udps(self, value: Optional[Sequence['GetTwingateResourceProtocolUdpArgs']]):
        pulumi.set(self, "udps", value)


@pulumi.input_type
class GetTwingateResourceProtocolTcpArgs:
    def __init__(__self__, *,
                 policy: str,
                 ports: Sequence[str]):
        pulumi.set(__self__, "policy", policy)
        pulumi.set(__self__, "ports", ports)

    @property
    @pulumi.getter
    def policy(self) -> str:
        return pulumi.get(self, "policy")

    @policy.setter
    def policy(self, value: str):
        pulumi.set(self, "policy", value)

    @property
    @pulumi.getter
    def ports(self) -> Sequence[str]:
        return pulumi.get(self, "ports")

    @ports.setter
    def ports(self, value: Sequence[str]):
        pulumi.set(self, "ports", value)


@pulumi.input_type
class GetTwingateResourceProtocolUdpArgs:
    def __init__(__self__, *,
                 policy: str,
                 ports: Sequence[str]):
        pulumi.set(__self__, "policy", policy)
        pulumi.set(__self__, "ports", ports)

    @property
    @pulumi.getter
    def policy(self) -> str:
        return pulumi.get(self, "policy")

    @policy.setter
    def policy(self, value: str):
        pulumi.set(self, "policy", value)

    @property
    @pulumi.getter
    def ports(self) -> Sequence[str]:
        return pulumi.get(self, "ports")

    @ports.setter
    def ports(self, value: Sequence[str]):
        pulumi.set(self, "ports", value)


@pulumi.input_type
class GetTwingateResourcesResourceArgs:
    def __init__(__self__, *,
                 address: str,
                 id: str,
                 name: str,
                 remote_network_id: str,
                 protocols: Optional[Sequence['GetTwingateResourcesResourceProtocolArgs']] = None):
        pulumi.set(__self__, "address", address)
        pulumi.set(__self__, "id", id)
        pulumi.set(__self__, "name", name)
        pulumi.set(__self__, "remote_network_id", remote_network_id)
        if protocols is not None:
            pulumi.set(__self__, "protocols", protocols)

    @property
    @pulumi.getter
    def address(self) -> str:
        return pulumi.get(self, "address")

    @address.setter
    def address(self, value: str):
        pulumi.set(self, "address", value)

    @property
    @pulumi.getter
    def id(self) -> str:
        return pulumi.get(self, "id")

    @id.setter
    def id(self, value: str):
        pulumi.set(self, "id", value)

    @property
    @pulumi.getter
    def name(self) -> str:
        return pulumi.get(self, "name")

    @name.setter
    def name(self, value: str):
        pulumi.set(self, "name", value)

    @property
    @pulumi.getter(name="remoteNetworkId")
    def remote_network_id(self) -> str:
        return pulumi.get(self, "remote_network_id")

    @remote_network_id.setter
    def remote_network_id(self, value: str):
        pulumi.set(self, "remote_network_id", value)

    @property
    @pulumi.getter
    def protocols(self) -> Optional[Sequence['GetTwingateResourcesResourceProtocolArgs']]:
        return pulumi.get(self, "protocols")

    @protocols.setter
    def protocols(self, value: Optional[Sequence['GetTwingateResourcesResourceProtocolArgs']]):
        pulumi.set(self, "protocols", value)


@pulumi.input_type
class GetTwingateResourcesResourceProtocolArgs:
    def __init__(__self__, *,
                 allow_icmp: bool,
                 tcps: Optional[Sequence['GetTwingateResourcesResourceProtocolTcpArgs']] = None,
                 udps: Optional[Sequence['GetTwingateResourcesResourceProtocolUdpArgs']] = None):
        pulumi.set(__self__, "allow_icmp", allow_icmp)
        if tcps is not None:
            pulumi.set(__self__, "tcps", tcps)
        if udps is not None:
            pulumi.set(__self__, "udps", udps)

    @property
    @pulumi.getter(name="allowIcmp")
    def allow_icmp(self) -> bool:
        return pulumi.get(self, "allow_icmp")

    @allow_icmp.setter
    def allow_icmp(self, value: bool):
        pulumi.set(self, "allow_icmp", value)

    @property
    @pulumi.getter
    def tcps(self) -> Optional[Sequence['GetTwingateResourcesResourceProtocolTcpArgs']]:
        return pulumi.get(self, "tcps")

    @tcps.setter
    def tcps(self, value: Optional[Sequence['GetTwingateResourcesResourceProtocolTcpArgs']]):
        pulumi.set(self, "tcps", value)

    @property
    @pulumi.getter
    def udps(self) -> Optional[Sequence['GetTwingateResourcesResourceProtocolUdpArgs']]:
        return pulumi.get(self, "udps")

    @udps.setter
    def udps(self, value: Optional[Sequence['GetTwingateResourcesResourceProtocolUdpArgs']]):
        pulumi.set(self, "udps", value)


@pulumi.input_type
class GetTwingateResourcesResourceProtocolTcpArgs:
    def __init__(__self__, *,
                 policy: str,
                 ports: Sequence[str]):
        pulumi.set(__self__, "policy", policy)
        pulumi.set(__self__, "ports", ports)

    @property
    @pulumi.getter
    def policy(self) -> str:
        return pulumi.get(self, "policy")

    @policy.setter
    def policy(self, value: str):
        pulumi.set(self, "policy", value)

    @property
    @pulumi.getter
    def ports(self) -> Sequence[str]:
        return pulumi.get(self, "ports")

    @ports.setter
    def ports(self, value: Sequence[str]):
        pulumi.set(self, "ports", value)


@pulumi.input_type
class GetTwingateResourcesResourceProtocolUdpArgs:
    def __init__(__self__, *,
                 policy: str,
                 ports: Sequence[str]):
        pulumi.set(__self__, "policy", policy)
        pulumi.set(__self__, "ports", ports)

    @property
    @pulumi.getter
    def policy(self) -> str:
        return pulumi.get(self, "policy")

    @policy.setter
    def policy(self, value: str):
        pulumi.set(self, "policy", value)

    @property
    @pulumi.getter
    def ports(self) -> Sequence[str]:
        return pulumi.get(self, "ports")

    @ports.setter
    def ports(self, value: Sequence[str]):
        pulumi.set(self, "ports", value)


@pulumi.input_type
class GetTwingateUsersUserArgs:
    def __init__(__self__, *,
                 email: str,
                 first_name: str,
                 id: str,
                 is_admin: bool,
                 last_name: str,
                 role: str):
        pulumi.set(__self__, "email", email)
        pulumi.set(__self__, "first_name", first_name)
        pulumi.set(__self__, "id", id)
        pulumi.set(__self__, "is_admin", is_admin)
        pulumi.set(__self__, "last_name", last_name)
        pulumi.set(__self__, "role", role)

    @property
    @pulumi.getter
    def email(self) -> str:
        return pulumi.get(self, "email")

    @email.setter
    def email(self, value: str):
        pulumi.set(self, "email", value)

    @property
    @pulumi.getter(name="firstName")
    def first_name(self) -> str:
        return pulumi.get(self, "first_name")

    @first_name.setter
    def first_name(self, value: str):
        pulumi.set(self, "first_name", value)

    @property
    @pulumi.getter
    def id(self) -> str:
        return pulumi.get(self, "id")

    @id.setter
    def id(self, value: str):
        pulumi.set(self, "id", value)

    @property
    @pulumi.getter(name="isAdmin")
    def is_admin(self) -> bool:
        return pulumi.get(self, "is_admin")

    @is_admin.setter
    def is_admin(self, value: bool):
        pulumi.set(self, "is_admin", value)

    @property
    @pulumi.getter(name="lastName")
    def last_name(self) -> str:
        return pulumi.get(self, "last_name")

    @last_name.setter
    def last_name(self, value: str):
        pulumi.set(self, "last_name", value)

    @property
    @pulumi.getter
    def role(self) -> str:
        return pulumi.get(self, "role")

    @role.setter
    def role(self, value: str):
        pulumi.set(self, "role", value)



# coding=utf-8
# *** WARNING: this file was generated by the Pulumi Terraform Bridge (tfgen) Tool. ***
# *** Do not edit by hand unless you're certain you know what you are doing! ***

import copy
import warnings
import pulumi
import pulumi.runtime
from typing import Any, Callable, Mapping, Optional, Sequence, Union, overload
from . import _utilities
from . import outputs
from ._inputs import *

__all__ = [
    'GetTwingateRemoteNetworksResult',
    'AwaitableGetTwingateRemoteNetworksResult',
    'get_twingate_remote_networks',
    'get_twingate_remote_networks_output',
]

@pulumi.output_type
class GetTwingateRemoteNetworksResult:
    """
    A collection of values returned by getTwingateRemoteNetworks.
    """
    def __init__(__self__, id=None, remote_networks=None):
        if id and not isinstance(id, str):
            raise TypeError("Expected argument 'id' to be a str")
        pulumi.set(__self__, "id", id)
        if remote_networks and not isinstance(remote_networks, list):
            raise TypeError("Expected argument 'remote_networks' to be a list")
        pulumi.set(__self__, "remote_networks", remote_networks)

    @property
    @pulumi.getter
    def id(self) -> str:
        """
        The provider-assigned unique ID for this managed resource.
        """
        return pulumi.get(self, "id")

    @property
    @pulumi.getter(name="remoteNetworks")
    def remote_networks(self) -> Optional[Sequence['outputs.GetTwingateRemoteNetworksRemoteNetworkResult']]:
        return pulumi.get(self, "remote_networks")


class AwaitableGetTwingateRemoteNetworksResult(GetTwingateRemoteNetworksResult):
    # pylint: disable=using-constant-test
    def __await__(self):
        if False:
            yield self
        return GetTwingateRemoteNetworksResult(
            id=self.id,
            remote_networks=self.remote_networks)


def get_twingate_remote_networks(remote_networks: Optional[Sequence[pulumi.InputType['GetTwingateRemoteNetworksRemoteNetworkArgs']]] = None,
                                 opts: Optional[pulumi.InvokeOptions] = None) -> AwaitableGetTwingateRemoteNetworksResult:
    """
    Use this data source to access information about an existing resource.
    """
    __args__ = dict()
    __args__['remoteNetworks'] = remote_networks
    opts = pulumi.InvokeOptions.merge(_utilities.get_invoke_opts_defaults(), opts)
    __ret__ = pulumi.runtime.invoke('twingate:index/getTwingateRemoteNetworks:getTwingateRemoteNetworks', __args__, opts=opts, typ=GetTwingateRemoteNetworksResult).value

    return AwaitableGetTwingateRemoteNetworksResult(
        id=pulumi.get(__ret__, 'id'),
        remote_networks=pulumi.get(__ret__, 'remote_networks'))


@_utilities.lift_output_func(get_twingate_remote_networks)
def get_twingate_remote_networks_output(remote_networks: Optional[pulumi.Input[Optional[Sequence[pulumi.InputType['GetTwingateRemoteNetworksRemoteNetworkArgs']]]]] = None,
                                        opts: Optional[pulumi.InvokeOptions] = None) -> pulumi.Output[GetTwingateRemoteNetworksResult]:
    """
    Use this data source to access information about an existing resource.
    """
    ...

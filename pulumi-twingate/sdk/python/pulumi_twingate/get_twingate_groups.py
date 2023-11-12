# coding=utf-8
# *** WARNING: this file was generated by the Pulumi Terraform Bridge (tfgen) Tool. ***
# *** Do not edit by hand unless you're certain you know what you are doing! ***

import copy
import warnings
import pulumi
import pulumi.runtime
from typing import Any, Mapping, Optional, Sequence, Union, overload
from . import _utilities
from . import outputs
from ._inputs import *

__all__ = [
    'GetTwingateGroupsResult',
    'AwaitableGetTwingateGroupsResult',
    'get_twingate_groups',
    'get_twingate_groups_output',
]

@pulumi.output_type
class GetTwingateGroupsResult:
    """
    A collection of values returned by getTwingateGroups.
    """
    def __init__(__self__, groups=None, id=None, is_active=None, name=None, type=None):
        if groups and not isinstance(groups, list):
            raise TypeError("Expected argument 'groups' to be a list")
        pulumi.set(__self__, "groups", groups)
        if id and not isinstance(id, str):
            raise TypeError("Expected argument 'id' to be a str")
        pulumi.set(__self__, "id", id)
        if is_active and not isinstance(is_active, bool):
            raise TypeError("Expected argument 'is_active' to be a bool")
        pulumi.set(__self__, "is_active", is_active)
        if name and not isinstance(name, str):
            raise TypeError("Expected argument 'name' to be a str")
        pulumi.set(__self__, "name", name)
        if type and not isinstance(type, str):
            raise TypeError("Expected argument 'type' to be a str")
        pulumi.set(__self__, "type", type)

    @property
    @pulumi.getter
    def groups(self) -> Sequence['outputs.GetTwingateGroupsGroupResult']:
        return pulumi.get(self, "groups")

    @property
    @pulumi.getter
    def id(self) -> str:
        return pulumi.get(self, "id")

    @property
    @pulumi.getter(name="isActive")
    def is_active(self) -> Optional[bool]:
        return pulumi.get(self, "is_active")

    @property
    @pulumi.getter
    def name(self) -> Optional[str]:
        return pulumi.get(self, "name")

    @property
    @pulumi.getter
    def type(self) -> Optional[str]:
        return pulumi.get(self, "type")


class AwaitableGetTwingateGroupsResult(GetTwingateGroupsResult):
    # pylint: disable=using-constant-test
    def __await__(self):
        if False:
            yield self
        return GetTwingateGroupsResult(
            groups=self.groups,
            id=self.id,
            is_active=self.is_active,
            name=self.name,
            type=self.type)


def get_twingate_groups(groups: Optional[Sequence[pulumi.InputType['GetTwingateGroupsGroupArgs']]] = None,
                        is_active: Optional[bool] = None,
                        name: Optional[str] = None,
                        type: Optional[str] = None,
                        opts: Optional[pulumi.InvokeOptions] = None) -> AwaitableGetTwingateGroupsResult:
    """
    Use this data source to access information about an existing resource.
    """
    __args__ = dict()
    __args__['groups'] = groups
    __args__['isActive'] = is_active
    __args__['name'] = name
    __args__['type'] = type
    opts = pulumi.InvokeOptions.merge(_utilities.get_invoke_opts_defaults(), opts)
    __ret__ = pulumi.runtime.invoke('twingate:index/getTwingateGroups:getTwingateGroups', __args__, opts=opts, typ=GetTwingateGroupsResult).value

    return AwaitableGetTwingateGroupsResult(
        groups=pulumi.get(__ret__, 'groups'),
        id=pulumi.get(__ret__, 'id'),
        is_active=pulumi.get(__ret__, 'is_active'),
        name=pulumi.get(__ret__, 'name'),
        type=pulumi.get(__ret__, 'type'))


@_utilities.lift_output_func(get_twingate_groups)
def get_twingate_groups_output(groups: Optional[pulumi.Input[Optional[Sequence[pulumi.InputType['GetTwingateGroupsGroupArgs']]]]] = None,
                               is_active: Optional[pulumi.Input[Optional[bool]]] = None,
                               name: Optional[pulumi.Input[Optional[str]]] = None,
                               type: Optional[pulumi.Input[Optional[str]]] = None,
                               opts: Optional[pulumi.InvokeOptions] = None) -> pulumi.Output[GetTwingateGroupsResult]:
    """
    Use this data source to access information about an existing resource.
    """
    ...

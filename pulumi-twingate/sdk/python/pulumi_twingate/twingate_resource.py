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

__all__ = ['TwingateResourceArgs', 'TwingateResource']

@pulumi.input_type
class TwingateResourceArgs:
    def __init__(__self__, *,
                 address: pulumi.Input[str],
                 remote_network_id: pulumi.Input[str],
                 access: Optional[pulumi.Input['TwingateResourceAccessArgs']] = None,
                 alias: Optional[pulumi.Input[str]] = None,
                 is_authoritative: Optional[pulumi.Input[bool]] = None,
                 is_browser_shortcut_enabled: Optional[pulumi.Input[bool]] = None,
                 is_visible: Optional[pulumi.Input[bool]] = None,
                 name: Optional[pulumi.Input[str]] = None,
                 protocols: Optional[pulumi.Input['TwingateResourceProtocolsArgs']] = None,
                 security_policy_id: Optional[pulumi.Input[str]] = None):
        """
        The set of arguments for constructing a TwingateResource resource.
        :param pulumi.Input[str] address: The Resource's IP/CIDR or FQDN/DNS zone
        :param pulumi.Input[str] remote_network_id: Remote Network ID where the Resource lives
        :param pulumi.Input['TwingateResourceAccessArgs'] access: Restrict access to certain groups or service accounts
        :param pulumi.Input[str] alias: Set a DNS alias address for the Resource. Must be a DNS-valid name string.
        :param pulumi.Input[bool] is_authoritative: Determines whether assignments in the access block will override any existing assignments. Default is `true`. If set to
               `false`, assignments made outside of Terraform will be ignored.
        :param pulumi.Input[bool] is_browser_shortcut_enabled: Controls whether an "Open in Browser" shortcut will be shown for this Resource in the Twingate Client.
        :param pulumi.Input[bool] is_visible: Controls whether this Resource will be visible in the main Resource list in the Twingate Client.
        :param pulumi.Input[str] name: The name of the Resource
        :param pulumi.Input['TwingateResourceProtocolsArgs'] protocols: Restrict access to certain protocols and ports. By default or when this argument is not defined, there is no
               restriction, and all protocols and ports are allowed.
        :param pulumi.Input[str] security_policy_id: The ID of a `twingate_security_policy` to set as this Resource's Security Policy.
        """
        pulumi.set(__self__, "address", address)
        pulumi.set(__self__, "remote_network_id", remote_network_id)
        if access is not None:
            pulumi.set(__self__, "access", access)
        if alias is not None:
            pulumi.set(__self__, "alias", alias)
        if is_authoritative is not None:
            pulumi.set(__self__, "is_authoritative", is_authoritative)
        if is_browser_shortcut_enabled is not None:
            pulumi.set(__self__, "is_browser_shortcut_enabled", is_browser_shortcut_enabled)
        if is_visible is not None:
            pulumi.set(__self__, "is_visible", is_visible)
        if name is not None:
            pulumi.set(__self__, "name", name)
        if protocols is not None:
            pulumi.set(__self__, "protocols", protocols)
        if security_policy_id is not None:
            pulumi.set(__self__, "security_policy_id", security_policy_id)

    @property
    @pulumi.getter
    def address(self) -> pulumi.Input[str]:
        """
        The Resource's IP/CIDR or FQDN/DNS zone
        """
        return pulumi.get(self, "address")

    @address.setter
    def address(self, value: pulumi.Input[str]):
        pulumi.set(self, "address", value)

    @property
    @pulumi.getter(name="remoteNetworkId")
    def remote_network_id(self) -> pulumi.Input[str]:
        """
        Remote Network ID where the Resource lives
        """
        return pulumi.get(self, "remote_network_id")

    @remote_network_id.setter
    def remote_network_id(self, value: pulumi.Input[str]):
        pulumi.set(self, "remote_network_id", value)

    @property
    @pulumi.getter
    def access(self) -> Optional[pulumi.Input['TwingateResourceAccessArgs']]:
        """
        Restrict access to certain groups or service accounts
        """
        return pulumi.get(self, "access")

    @access.setter
    def access(self, value: Optional[pulumi.Input['TwingateResourceAccessArgs']]):
        pulumi.set(self, "access", value)

    @property
    @pulumi.getter
    def alias(self) -> Optional[pulumi.Input[str]]:
        """
        Set a DNS alias address for the Resource. Must be a DNS-valid name string.
        """
        return pulumi.get(self, "alias")

    @alias.setter
    def alias(self, value: Optional[pulumi.Input[str]]):
        pulumi.set(self, "alias", value)

    @property
    @pulumi.getter(name="isAuthoritative")
    def is_authoritative(self) -> Optional[pulumi.Input[bool]]:
        """
        Determines whether assignments in the access block will override any existing assignments. Default is `true`. If set to
        `false`, assignments made outside of Terraform will be ignored.
        """
        return pulumi.get(self, "is_authoritative")

    @is_authoritative.setter
    def is_authoritative(self, value: Optional[pulumi.Input[bool]]):
        pulumi.set(self, "is_authoritative", value)

    @property
    @pulumi.getter(name="isBrowserShortcutEnabled")
    def is_browser_shortcut_enabled(self) -> Optional[pulumi.Input[bool]]:
        """
        Controls whether an "Open in Browser" shortcut will be shown for this Resource in the Twingate Client.
        """
        return pulumi.get(self, "is_browser_shortcut_enabled")

    @is_browser_shortcut_enabled.setter
    def is_browser_shortcut_enabled(self, value: Optional[pulumi.Input[bool]]):
        pulumi.set(self, "is_browser_shortcut_enabled", value)

    @property
    @pulumi.getter(name="isVisible")
    def is_visible(self) -> Optional[pulumi.Input[bool]]:
        """
        Controls whether this Resource will be visible in the main Resource list in the Twingate Client.
        """
        return pulumi.get(self, "is_visible")

    @is_visible.setter
    def is_visible(self, value: Optional[pulumi.Input[bool]]):
        pulumi.set(self, "is_visible", value)

    @property
    @pulumi.getter
    def name(self) -> Optional[pulumi.Input[str]]:
        """
        The name of the Resource
        """
        return pulumi.get(self, "name")

    @name.setter
    def name(self, value: Optional[pulumi.Input[str]]):
        pulumi.set(self, "name", value)

    @property
    @pulumi.getter
    def protocols(self) -> Optional[pulumi.Input['TwingateResourceProtocolsArgs']]:
        """
        Restrict access to certain protocols and ports. By default or when this argument is not defined, there is no
        restriction, and all protocols and ports are allowed.
        """
        return pulumi.get(self, "protocols")

    @protocols.setter
    def protocols(self, value: Optional[pulumi.Input['TwingateResourceProtocolsArgs']]):
        pulumi.set(self, "protocols", value)

    @property
    @pulumi.getter(name="securityPolicyId")
    def security_policy_id(self) -> Optional[pulumi.Input[str]]:
        """
        The ID of a `twingate_security_policy` to set as this Resource's Security Policy.
        """
        return pulumi.get(self, "security_policy_id")

    @security_policy_id.setter
    def security_policy_id(self, value: Optional[pulumi.Input[str]]):
        pulumi.set(self, "security_policy_id", value)


@pulumi.input_type
class _TwingateResourceState:
    def __init__(__self__, *,
                 access: Optional[pulumi.Input['TwingateResourceAccessArgs']] = None,
                 address: Optional[pulumi.Input[str]] = None,
                 alias: Optional[pulumi.Input[str]] = None,
                 is_authoritative: Optional[pulumi.Input[bool]] = None,
                 is_browser_shortcut_enabled: Optional[pulumi.Input[bool]] = None,
                 is_visible: Optional[pulumi.Input[bool]] = None,
                 name: Optional[pulumi.Input[str]] = None,
                 protocols: Optional[pulumi.Input['TwingateResourceProtocolsArgs']] = None,
                 remote_network_id: Optional[pulumi.Input[str]] = None,
                 security_policy_id: Optional[pulumi.Input[str]] = None):
        """
        Input properties used for looking up and filtering TwingateResource resources.
        :param pulumi.Input['TwingateResourceAccessArgs'] access: Restrict access to certain groups or service accounts
        :param pulumi.Input[str] address: The Resource's IP/CIDR or FQDN/DNS zone
        :param pulumi.Input[str] alias: Set a DNS alias address for the Resource. Must be a DNS-valid name string.
        :param pulumi.Input[bool] is_authoritative: Determines whether assignments in the access block will override any existing assignments. Default is `true`. If set to
               `false`, assignments made outside of Terraform will be ignored.
        :param pulumi.Input[bool] is_browser_shortcut_enabled: Controls whether an "Open in Browser" shortcut will be shown for this Resource in the Twingate Client.
        :param pulumi.Input[bool] is_visible: Controls whether this Resource will be visible in the main Resource list in the Twingate Client.
        :param pulumi.Input[str] name: The name of the Resource
        :param pulumi.Input['TwingateResourceProtocolsArgs'] protocols: Restrict access to certain protocols and ports. By default or when this argument is not defined, there is no
               restriction, and all protocols and ports are allowed.
        :param pulumi.Input[str] remote_network_id: Remote Network ID where the Resource lives
        :param pulumi.Input[str] security_policy_id: The ID of a `twingate_security_policy` to set as this Resource's Security Policy.
        """
        if access is not None:
            pulumi.set(__self__, "access", access)
        if address is not None:
            pulumi.set(__self__, "address", address)
        if alias is not None:
            pulumi.set(__self__, "alias", alias)
        if is_authoritative is not None:
            pulumi.set(__self__, "is_authoritative", is_authoritative)
        if is_browser_shortcut_enabled is not None:
            pulumi.set(__self__, "is_browser_shortcut_enabled", is_browser_shortcut_enabled)
        if is_visible is not None:
            pulumi.set(__self__, "is_visible", is_visible)
        if name is not None:
            pulumi.set(__self__, "name", name)
        if protocols is not None:
            pulumi.set(__self__, "protocols", protocols)
        if remote_network_id is not None:
            pulumi.set(__self__, "remote_network_id", remote_network_id)
        if security_policy_id is not None:
            pulumi.set(__self__, "security_policy_id", security_policy_id)

    @property
    @pulumi.getter
    def access(self) -> Optional[pulumi.Input['TwingateResourceAccessArgs']]:
        """
        Restrict access to certain groups or service accounts
        """
        return pulumi.get(self, "access")

    @access.setter
    def access(self, value: Optional[pulumi.Input['TwingateResourceAccessArgs']]):
        pulumi.set(self, "access", value)

    @property
    @pulumi.getter
    def address(self) -> Optional[pulumi.Input[str]]:
        """
        The Resource's IP/CIDR or FQDN/DNS zone
        """
        return pulumi.get(self, "address")

    @address.setter
    def address(self, value: Optional[pulumi.Input[str]]):
        pulumi.set(self, "address", value)

    @property
    @pulumi.getter
    def alias(self) -> Optional[pulumi.Input[str]]:
        """
        Set a DNS alias address for the Resource. Must be a DNS-valid name string.
        """
        return pulumi.get(self, "alias")

    @alias.setter
    def alias(self, value: Optional[pulumi.Input[str]]):
        pulumi.set(self, "alias", value)

    @property
    @pulumi.getter(name="isAuthoritative")
    def is_authoritative(self) -> Optional[pulumi.Input[bool]]:
        """
        Determines whether assignments in the access block will override any existing assignments. Default is `true`. If set to
        `false`, assignments made outside of Terraform will be ignored.
        """
        return pulumi.get(self, "is_authoritative")

    @is_authoritative.setter
    def is_authoritative(self, value: Optional[pulumi.Input[bool]]):
        pulumi.set(self, "is_authoritative", value)

    @property
    @pulumi.getter(name="isBrowserShortcutEnabled")
    def is_browser_shortcut_enabled(self) -> Optional[pulumi.Input[bool]]:
        """
        Controls whether an "Open in Browser" shortcut will be shown for this Resource in the Twingate Client.
        """
        return pulumi.get(self, "is_browser_shortcut_enabled")

    @is_browser_shortcut_enabled.setter
    def is_browser_shortcut_enabled(self, value: Optional[pulumi.Input[bool]]):
        pulumi.set(self, "is_browser_shortcut_enabled", value)

    @property
    @pulumi.getter(name="isVisible")
    def is_visible(self) -> Optional[pulumi.Input[bool]]:
        """
        Controls whether this Resource will be visible in the main Resource list in the Twingate Client.
        """
        return pulumi.get(self, "is_visible")

    @is_visible.setter
    def is_visible(self, value: Optional[pulumi.Input[bool]]):
        pulumi.set(self, "is_visible", value)

    @property
    @pulumi.getter
    def name(self) -> Optional[pulumi.Input[str]]:
        """
        The name of the Resource
        """
        return pulumi.get(self, "name")

    @name.setter
    def name(self, value: Optional[pulumi.Input[str]]):
        pulumi.set(self, "name", value)

    @property
    @pulumi.getter
    def protocols(self) -> Optional[pulumi.Input['TwingateResourceProtocolsArgs']]:
        """
        Restrict access to certain protocols and ports. By default or when this argument is not defined, there is no
        restriction, and all protocols and ports are allowed.
        """
        return pulumi.get(self, "protocols")

    @protocols.setter
    def protocols(self, value: Optional[pulumi.Input['TwingateResourceProtocolsArgs']]):
        pulumi.set(self, "protocols", value)

    @property
    @pulumi.getter(name="remoteNetworkId")
    def remote_network_id(self) -> Optional[pulumi.Input[str]]:
        """
        Remote Network ID where the Resource lives
        """
        return pulumi.get(self, "remote_network_id")

    @remote_network_id.setter
    def remote_network_id(self, value: Optional[pulumi.Input[str]]):
        pulumi.set(self, "remote_network_id", value)

    @property
    @pulumi.getter(name="securityPolicyId")
    def security_policy_id(self) -> Optional[pulumi.Input[str]]:
        """
        The ID of a `twingate_security_policy` to set as this Resource's Security Policy.
        """
        return pulumi.get(self, "security_policy_id")

    @security_policy_id.setter
    def security_policy_id(self, value: Optional[pulumi.Input[str]]):
        pulumi.set(self, "security_policy_id", value)


class TwingateResource(pulumi.CustomResource):
    @overload
    def __init__(__self__,
                 resource_name: str,
                 opts: Optional[pulumi.ResourceOptions] = None,
                 access: Optional[pulumi.Input[pulumi.InputType['TwingateResourceAccessArgs']]] = None,
                 address: Optional[pulumi.Input[str]] = None,
                 alias: Optional[pulumi.Input[str]] = None,
                 is_authoritative: Optional[pulumi.Input[bool]] = None,
                 is_browser_shortcut_enabled: Optional[pulumi.Input[bool]] = None,
                 is_visible: Optional[pulumi.Input[bool]] = None,
                 name: Optional[pulumi.Input[str]] = None,
                 protocols: Optional[pulumi.Input[pulumi.InputType['TwingateResourceProtocolsArgs']]] = None,
                 remote_network_id: Optional[pulumi.Input[str]] = None,
                 security_policy_id: Optional[pulumi.Input[str]] = None,
                 __props__=None):
        """
        Create a TwingateResource resource with the given unique name, props, and options.
        :param str resource_name: The name of the resource.
        :param pulumi.ResourceOptions opts: Options for the resource.
        :param pulumi.Input[pulumi.InputType['TwingateResourceAccessArgs']] access: Restrict access to certain groups or service accounts
        :param pulumi.Input[str] address: The Resource's IP/CIDR or FQDN/DNS zone
        :param pulumi.Input[str] alias: Set a DNS alias address for the Resource. Must be a DNS-valid name string.
        :param pulumi.Input[bool] is_authoritative: Determines whether assignments in the access block will override any existing assignments. Default is `true`. If set to
               `false`, assignments made outside of Terraform will be ignored.
        :param pulumi.Input[bool] is_browser_shortcut_enabled: Controls whether an "Open in Browser" shortcut will be shown for this Resource in the Twingate Client.
        :param pulumi.Input[bool] is_visible: Controls whether this Resource will be visible in the main Resource list in the Twingate Client.
        :param pulumi.Input[str] name: The name of the Resource
        :param pulumi.Input[pulumi.InputType['TwingateResourceProtocolsArgs']] protocols: Restrict access to certain protocols and ports. By default or when this argument is not defined, there is no
               restriction, and all protocols and ports are allowed.
        :param pulumi.Input[str] remote_network_id: Remote Network ID where the Resource lives
        :param pulumi.Input[str] security_policy_id: The ID of a `twingate_security_policy` to set as this Resource's Security Policy.
        """
        ...
    @overload
    def __init__(__self__,
                 resource_name: str,
                 args: TwingateResourceArgs,
                 opts: Optional[pulumi.ResourceOptions] = None):
        """
        Create a TwingateResource resource with the given unique name, props, and options.
        :param str resource_name: The name of the resource.
        :param TwingateResourceArgs args: The arguments to use to populate this resource's properties.
        :param pulumi.ResourceOptions opts: Options for the resource.
        """
        ...
    def __init__(__self__, resource_name: str, *args, **kwargs):
        resource_args, opts = _utilities.get_resource_args_opts(TwingateResourceArgs, pulumi.ResourceOptions, *args, **kwargs)
        if resource_args is not None:
            __self__._internal_init(resource_name, opts, **resource_args.__dict__)
        else:
            __self__._internal_init(resource_name, *args, **kwargs)

    def _internal_init(__self__,
                 resource_name: str,
                 opts: Optional[pulumi.ResourceOptions] = None,
                 access: Optional[pulumi.Input[pulumi.InputType['TwingateResourceAccessArgs']]] = None,
                 address: Optional[pulumi.Input[str]] = None,
                 alias: Optional[pulumi.Input[str]] = None,
                 is_authoritative: Optional[pulumi.Input[bool]] = None,
                 is_browser_shortcut_enabled: Optional[pulumi.Input[bool]] = None,
                 is_visible: Optional[pulumi.Input[bool]] = None,
                 name: Optional[pulumi.Input[str]] = None,
                 protocols: Optional[pulumi.Input[pulumi.InputType['TwingateResourceProtocolsArgs']]] = None,
                 remote_network_id: Optional[pulumi.Input[str]] = None,
                 security_policy_id: Optional[pulumi.Input[str]] = None,
                 __props__=None):
        opts = pulumi.ResourceOptions.merge(_utilities.get_resource_opts_defaults(), opts)
        if not isinstance(opts, pulumi.ResourceOptions):
            raise TypeError('Expected resource options to be a ResourceOptions instance')
        if opts.id is None:
            if __props__ is not None:
                raise TypeError('__props__ is only valid when passed in combination with a valid opts.id to get an existing resource')
            __props__ = TwingateResourceArgs.__new__(TwingateResourceArgs)

            __props__.__dict__["access"] = access
            if address is None and not opts.urn:
                raise TypeError("Missing required property 'address'")
            __props__.__dict__["address"] = address
            __props__.__dict__["alias"] = alias
            __props__.__dict__["is_authoritative"] = is_authoritative
            __props__.__dict__["is_browser_shortcut_enabled"] = is_browser_shortcut_enabled
            __props__.__dict__["is_visible"] = is_visible
            __props__.__dict__["name"] = name
            __props__.__dict__["protocols"] = protocols
            if remote_network_id is None and not opts.urn:
                raise TypeError("Missing required property 'remote_network_id'")
            __props__.__dict__["remote_network_id"] = remote_network_id
            __props__.__dict__["security_policy_id"] = security_policy_id
        super(TwingateResource, __self__).__init__(
            'twingate:index/twingateResource:TwingateResource',
            resource_name,
            __props__,
            opts)

    @staticmethod
    def get(resource_name: str,
            id: pulumi.Input[str],
            opts: Optional[pulumi.ResourceOptions] = None,
            access: Optional[pulumi.Input[pulumi.InputType['TwingateResourceAccessArgs']]] = None,
            address: Optional[pulumi.Input[str]] = None,
            alias: Optional[pulumi.Input[str]] = None,
            is_authoritative: Optional[pulumi.Input[bool]] = None,
            is_browser_shortcut_enabled: Optional[pulumi.Input[bool]] = None,
            is_visible: Optional[pulumi.Input[bool]] = None,
            name: Optional[pulumi.Input[str]] = None,
            protocols: Optional[pulumi.Input[pulumi.InputType['TwingateResourceProtocolsArgs']]] = None,
            remote_network_id: Optional[pulumi.Input[str]] = None,
            security_policy_id: Optional[pulumi.Input[str]] = None) -> 'TwingateResource':
        """
        Get an existing TwingateResource resource's state with the given name, id, and optional extra
        properties used to qualify the lookup.

        :param str resource_name: The unique name of the resulting resource.
        :param pulumi.Input[str] id: The unique provider ID of the resource to lookup.
        :param pulumi.ResourceOptions opts: Options for the resource.
        :param pulumi.Input[pulumi.InputType['TwingateResourceAccessArgs']] access: Restrict access to certain groups or service accounts
        :param pulumi.Input[str] address: The Resource's IP/CIDR or FQDN/DNS zone
        :param pulumi.Input[str] alias: Set a DNS alias address for the Resource. Must be a DNS-valid name string.
        :param pulumi.Input[bool] is_authoritative: Determines whether assignments in the access block will override any existing assignments. Default is `true`. If set to
               `false`, assignments made outside of Terraform will be ignored.
        :param pulumi.Input[bool] is_browser_shortcut_enabled: Controls whether an "Open in Browser" shortcut will be shown for this Resource in the Twingate Client.
        :param pulumi.Input[bool] is_visible: Controls whether this Resource will be visible in the main Resource list in the Twingate Client.
        :param pulumi.Input[str] name: The name of the Resource
        :param pulumi.Input[pulumi.InputType['TwingateResourceProtocolsArgs']] protocols: Restrict access to certain protocols and ports. By default or when this argument is not defined, there is no
               restriction, and all protocols and ports are allowed.
        :param pulumi.Input[str] remote_network_id: Remote Network ID where the Resource lives
        :param pulumi.Input[str] security_policy_id: The ID of a `twingate_security_policy` to set as this Resource's Security Policy.
        """
        opts = pulumi.ResourceOptions.merge(opts, pulumi.ResourceOptions(id=id))

        __props__ = _TwingateResourceState.__new__(_TwingateResourceState)

        __props__.__dict__["access"] = access
        __props__.__dict__["address"] = address
        __props__.__dict__["alias"] = alias
        __props__.__dict__["is_authoritative"] = is_authoritative
        __props__.__dict__["is_browser_shortcut_enabled"] = is_browser_shortcut_enabled
        __props__.__dict__["is_visible"] = is_visible
        __props__.__dict__["name"] = name
        __props__.__dict__["protocols"] = protocols
        __props__.__dict__["remote_network_id"] = remote_network_id
        __props__.__dict__["security_policy_id"] = security_policy_id
        return TwingateResource(resource_name, opts=opts, __props__=__props__)

    @property
    @pulumi.getter
    def access(self) -> pulumi.Output[Optional['outputs.TwingateResourceAccess']]:
        """
        Restrict access to certain groups or service accounts
        """
        return pulumi.get(self, "access")

    @property
    @pulumi.getter
    def address(self) -> pulumi.Output[str]:
        """
        The Resource's IP/CIDR or FQDN/DNS zone
        """
        return pulumi.get(self, "address")

    @property
    @pulumi.getter
    def alias(self) -> pulumi.Output[Optional[str]]:
        """
        Set a DNS alias address for the Resource. Must be a DNS-valid name string.
        """
        return pulumi.get(self, "alias")

    @property
    @pulumi.getter(name="isAuthoritative")
    def is_authoritative(self) -> pulumi.Output[bool]:
        """
        Determines whether assignments in the access block will override any existing assignments. Default is `true`. If set to
        `false`, assignments made outside of Terraform will be ignored.
        """
        return pulumi.get(self, "is_authoritative")

    @property
    @pulumi.getter(name="isBrowserShortcutEnabled")
    def is_browser_shortcut_enabled(self) -> pulumi.Output[bool]:
        """
        Controls whether an "Open in Browser" shortcut will be shown for this Resource in the Twingate Client.
        """
        return pulumi.get(self, "is_browser_shortcut_enabled")

    @property
    @pulumi.getter(name="isVisible")
    def is_visible(self) -> pulumi.Output[bool]:
        """
        Controls whether this Resource will be visible in the main Resource list in the Twingate Client.
        """
        return pulumi.get(self, "is_visible")

    @property
    @pulumi.getter
    def name(self) -> pulumi.Output[str]:
        """
        The name of the Resource
        """
        return pulumi.get(self, "name")

    @property
    @pulumi.getter
    def protocols(self) -> pulumi.Output['outputs.TwingateResourceProtocols']:
        """
        Restrict access to certain protocols and ports. By default or when this argument is not defined, there is no
        restriction, and all protocols and ports are allowed.
        """
        return pulumi.get(self, "protocols")

    @property
    @pulumi.getter(name="remoteNetworkId")
    def remote_network_id(self) -> pulumi.Output[str]:
        """
        Remote Network ID where the Resource lives
        """
        return pulumi.get(self, "remote_network_id")

    @property
    @pulumi.getter(name="securityPolicyId")
    def security_policy_id(self) -> pulumi.Output[str]:
        """
        The ID of a `twingate_security_policy` to set as this Resource's Security Policy.
        """
        return pulumi.get(self, "security_policy_id")


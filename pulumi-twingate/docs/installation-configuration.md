---
title: Twingate Installation & Configuration
meta_desc: Information on how to install the Twingate provider.
layout: package
---

## Installation

The Pulumi Twingate provider is available as a package in all Pulumi languages:

* JavaScript/TypeScript: [`@twingate-labs/pulumi-twingate`](https://www.npmjs.com/package/@twingate-labs/pulumi-twingate)
* Python: [`pulumi_twingate`](https://pypi.org/project/pulumi-twingate/)
* Go: [`github.com/Twingate-Labs/pulumi-twingate/sdk/go/twingate`](https://github.com/Twingate-Labs/pulumi-twingate/tree/master/sdk/go/twingate)
* .NET: [`TwingateLabs.Twingate`](https://www.nuget.org/packages/TwingateLabs.Twingate)

### Provider Binary

The Twingate provider binary is a third party binary, this should be installed automatically by package managers (e.g. Python Pip). It can be manually installed using the `pulumi plugin` command.

```bash
pulumi plugin install resource twingate --server github://api.github.com/twingate-labs
```

Replace the version string with your desired version.

## Setup

To provision resources with the Pulumi Twingate provider, you need to have Twingate credentials.

### Set environment variables

Once you have provisioned these credentials, you can set environment variables to provision resources in Twingate:

{{< chooser os "linux,macos,windows" >}}
{{% choosable os linux %}}

```bash
$ export TWINGATE_API_TOKEN=<TWINGATE_API_TOKEN>
$ export TWINGATE_NETWORK=<TWINGATE_NETWORK>
```

{{% /choosable %}}

{{% choosable os macos %}}

```bash
$ export TWINGATE_API_TOKEN=<TWINGATE_API_TOKEN>
$ export TWINGATE_NETWORK=<TWINGATE_NETWORK>
```

{{% /choosable %}}

{{% choosable os windows %}}

```powershell
> $env:TWINGATE_API_TOKEN = "<TWINGATE_API_TOKEN>"
> $env:TWINGATE_NETWORK = "<TWINGATE_NETWORK>"
```

{{% /choosable %}}
{{< /chooser >}}

If you prefer that they be stored alongside your Pulumi stack for easy multi-user access:

    ```bash
    $ pulumi config set twingate:apiToken XXXXXX --secret
    $ pulumi config set twingate:network YYYYYY
    ```

The complete list of
configuration parameters is in the [Twingate provider README](https://github.com/Twingate-Labs/pulumi-twingate/blob/master/README.md).

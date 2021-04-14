
Twingate Terraform Provider
==================

- Website: https://www.terraform.io
- [![Gitter chat](https://badges.gitter.im/hashicorp-terraform/Lobby.png)](https://gitter.im/hashicorp-terraform/Lobby)
- [![Coverage Status](https://coveralls.io/repos/github/Twingate/terraform-provider-twingate/badge.svg?branch=feature/TG-2579-ci-and-initital-setup&t=rqgifB)](https://coveralls.io/github/Twingate/terraform-provider-twingate?branch=feature/TG-2579-ci-and-initital-setup)
- Mailing list: [Google Groups](http://groups.google.com/group/terraform-tool)

<img src="https://cdn.rawgit.com/hashicorp/terraform-website/master/content/source/assets/images/logo-hashicorp.svg" width="600px">

Requirements
------------

-	[Terraform](https://www.terraform.io/downloads.html) 0.14.x
-	[Go](https://golang.org/doc/install) 1.16.2 (to build the provider plugin)

## Build: 

Run the following command to build the provider

```shell
make build
```

## Install:

First, build and install the provider.

```shell
make install
```
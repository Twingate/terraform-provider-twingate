
Twingate Terraform Provider
==================



[![Coverage Status](https://coveralls.io/repos/github/Twingate/terraform-provider-twingate/badge.svg?branch=main&t=rqgifB)](https://coveralls.io/github/Twingate/terraform-provider-twingate?branch=main)

Requirements
------------

-	[Terraform](https://www.terraform.io/downloads.html) 0.15.x
-	[Go](https://golang.org/doc/install) 1.16.3 (to build the provider plugin)

## Build: 

Run the following command to build the provider

```shell
make build
```

## Install:

Install the provider for local testing.

```shell
make install
```

## Documentation:

To update the documentation edit the files in `templates/` and then run `make docs`.  The files in `docs/` are auto-generated and should not be updated manually.

## Contributions:

Contributions to this project are [released](https://help.github.com/articles/github-terms-of-service/#6-contributions-under-repository-license) under the [project's open source license](LICENSE).

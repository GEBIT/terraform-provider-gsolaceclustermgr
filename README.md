# TODO rewrite this!

# Terraform Provider for Solace MissionControl ClusterManager

THis provider is based on the [HashiCorp Developer Tutorial](https://developer.hashicorp.com/terraform/tutorials/providers-plugin-framework)


Please see the [GitHub template repository documentation](https://help.github.com/en/github/creating-cloning-and-archiving-repositories/creating-a-repository-from-a-template) for how to create a new repository from this template on GitHub.

TODO o [publish it on the Terraform Registry](https://developer.hashicorp.com/terraform/registry/providers/publishing) so that others can use it.

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.21

## Building The Provider

1. Clone the repository
1. Enter the repository directory
1. Build the provider using the Go `install` command:

```shell
go install
```

## Adding Dependencies

This provider uses [Go modules](https://github.com/golang/go/wiki/Modules).
Please see the Go documentation for the most up to date information about using Go modules.

```shell
go get github.com/author/dependency
go mod tidy
```

Then commit the changes to `go.mod` and `go.sum`.

## Using the provider

The provider allos you to create/update/delate Solace cloud API broker instances.

Define the provider with the solace API URL and a valid bearerToken:
~~~
provider "gsolaceclustermgr" {
  bearer_token = "<someBearerToken>"
  host = "https://api.solace.cloud"
}
~~~
Then create a broker using the *gsolaceclustermgr_broker* resource
~~~
resource "gsolaceclustermgr_broker" "ocs-test" {
  count           = 1
  serviceclass_id = "ENTERPRISE_250_STANDALONE"
  name            = "ocs-prov-testHH"
  datacenter_id   = "aks-germanywestcentral"
  msg_vpn_name    = "ocs-msgvpn-1"
  cluster_name    = "gwc-aks-cluster1"
}
~~~
Updating the broker is supported - but only the name attribute may be changed.

The offical solace terraform provider should cover further manipulation like messageVPN setup.

## Provider Implementation

This provider only supports a small part of the missoncontrol API v2, namely those that are needed to  create and delete solace brokers in the solace cloud.


The REST client to access the API is generated using github.com/deepmap/oapi-codegen. 
For CI testing the provider without actually calling the productive solace API, a fakeserver is included.

This is the first project I've done in go - improvements and feedback are welcome.

## Developing the Provider

If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed on your machine (see [Requirements](#requirements) above).

To compile the provider, run `go install`. This will build the provider and put the provider binary in the `$GOPATH/bin` directory.

To generate or update documentation, run `go generate`.

In order to run the full suite of Acceptance tests, run `make testacc`.

*Note:* Acceptance tests create real resources, and often cost money to run.

```shell
make testacc
```

for testing with terraform put this into your `%APPDATA%\terraform.rc` file:
~~~
provider_installation {

  dev_overrides {
	  "gebit.de/tf/gsolaceclustermgr" = "C:/Users/<you>/go/bin"
  }
  direct {}
}
~~~
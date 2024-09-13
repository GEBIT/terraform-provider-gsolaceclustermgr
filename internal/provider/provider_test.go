package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

const (
	// shared test providerConfig using fakeserver
	providerConfig = `
	provider "gsolaceclustermgr" {
		bearer_token = "bt42"	
		host = "http://localhost:8091"
		}
		`
)

var (
	// testAccProtoV6ProviderFactories are used to instantiate a provider during
	// acceptance testing. The factory function will be invoked for every Terraform
	// CLI command executed to create a provider server to which the CLI can
	// reattach.
	testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
		"gsolaceclustermgr": providerserver.NewProtocol6WithError(New("test")()),
	}
)

package provider

import (
	"fmt"
	"os"
	"terraform-provider-gsolaceclustermgr/internal/fakeserver"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

var svr *fakeserver.Fakeserver

func startFakeServer() {
	apiServerObjects := make(map[string]fakeserver.ServiceInfo)
	port := 8091
	fmt.Printf("Starting fake server on port %d...\n", port)
	svr = fakeserver.NewFakeServer(port, apiServerObjects, true, os.Getenv("DEBUG") != "")
}

func stopFakeServer() {
	if svr != nil {
		fmt.Printf("Shutting down fake server ...\n")
		svr.Shutdown()
	}
}

func TestAccBrokerResource(t *testing.T) {
	if os.Getenv("EXT_SERVER") == "" {
		startFakeServer()
		defer stopFakeServer()
	}
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig + `
resource "gsolaceclustermgr_broker" "test" {
	serviceclass_id = "ENTERPRISE_250_STANDALONE"
	name            = "ocs-prov-test"
	datacenter_id   = "aks-germanywestcentral"
	msg_vpn_name    = "ocs-msgvpn-1"
	cluster_name    = "gwc-aks-cluster1"
}
`,
				ConfigStateChecks: []statecheck.StateCheck{
					// verify attributes
					statecheck.ExpectKnownValue(
						"gsolaceclustermgr_broker.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact("ocs-prov-test"),
					),
					statecheck.ExpectKnownValue(
						"gsolaceclustermgr_broker.test",
						tfjsonpath.New("serviceclass_id"),
						knownvalue.StringExact("ENTERPRISE_250_STANDALONE"),
					),
					statecheck.ExpectKnownValue(
						"gsolaceclustermgr_broker.test",
						tfjsonpath.New("datacenter_id"),
						knownvalue.StringExact("aks-germanywestcentral"),
					),
					statecheck.ExpectKnownValue(
						"gsolaceclustermgr_broker.test",
						tfjsonpath.New("msg_vpn_name"),
						knownvalue.StringExact("ocs-msgvpn-1"),
					),
					statecheck.ExpectKnownValue(
						"gsolaceclustermgr_broker.test",
						tfjsonpath.New("cluster_name"),
						knownvalue.StringExact("gwc-aks-cluster1"),
					),
					// Verify Computed attributes
					statecheck.ExpectKnownValue(
						"gsolaceclustermgr_broker.test",
						tfjsonpath.New("status"),
						knownvalue.StringExact("COMPLETED"),
					),
					// Verify dynamic values have any value set in the state.
					statecheck.ExpectKnownValue(
						"gsolaceclustermgr_broker.test",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"gsolaceclustermgr_broker.test",
						tfjsonpath.New("last_updated"),
						knownvalue.NotNull(),
					),
				},
			},
			// Update and Read testing
			{
				Config: providerConfig + `
resource "gsolaceclustermgr_broker" "test" {
	serviceclass_id = "ENTERPRISE_250_STANDALONE"
	name            = "ocs-prov-test-changed"
	datacenter_id   = "aks-germanywestcentral"
	msg_vpn_name    = "ocs-msgvpn-1"
	cluster_name    = "gwc-aks-cluster1"
}
`,
				ExpectNonEmptyPlan: true,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"gsolaceclustermgr_broker.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact("ocs-prov-test-changed"),
					), statecheck.ExpectKnownValue(
						"gsolaceclustermgr_broker.test",
						tfjsonpath.New("last_updated"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"gsolaceclustermgr_broker.test",
						tfjsonpath.New("status"),
						knownvalue.StringExact("PENDING"),
					),
				},
			},

			// Delete testing automatically occurs in TestCase
		},
	})
}

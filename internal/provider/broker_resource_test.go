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
			{
				Config: testConfig("test", "ocs-prov-test", true),
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
					// verify optional attributes
					statecheck.ExpectKnownValue(
						"gsolaceclustermgr_broker.test",
						tfjsonpath.New("msg_vpn_name"),
						knownvalue.StringExact("ocs-msgvpn"),
					),
					statecheck.ExpectKnownValue(
						"gsolaceclustermgr_broker.test",
						tfjsonpath.New("cluster_name"),
						knownvalue.StringExact("gwc-aks-ocs"),
					),
					statecheck.ExpectKnownValue(
						"gsolaceclustermgr_broker.test",
						tfjsonpath.New("custom_router_name"),
						knownvalue.StringExact("ocs-router"),
					),
					statecheck.ExpectKnownValue(
						"gsolaceclustermgr_broker.test",
						tfjsonpath.New("event_broker_version"),
						knownvalue.StringExact("1.2.3"),
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
			// Delete testing automatically occurs in TestCase
		},
	})
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing (optionals not set)
			{
				Config: testConfig("test2", "ocs-prov-test2", false),
				ConfigStateChecks: []statecheck.StateCheck{
					// verify attributes
					statecheck.ExpectKnownValue(
						"gsolaceclustermgr_broker.test2",
						tfjsonpath.New("name"),
						knownvalue.StringExact("ocs-prov-test2"),
					),
					statecheck.ExpectKnownValue(
						"gsolaceclustermgr_broker.test2",
						tfjsonpath.New("serviceclass_id"),
						knownvalue.StringExact("ENTERPRISE_250_STANDALONE"),
					),
					statecheck.ExpectKnownValue(
						"gsolaceclustermgr_broker.test2",
						tfjsonpath.New("datacenter_id"),
						knownvalue.StringExact("aks-germanywestcentral"),
					),
					// verify optional attributes mock default values
					statecheck.ExpectKnownValue(
						"gsolaceclustermgr_broker.test2",
						tfjsonpath.New("msg_vpn_name"),
						knownvalue.StringExact("test-vpn1"),
					),
					statecheck.ExpectKnownValue(
						"gsolaceclustermgr_broker.test2",
						tfjsonpath.New("cluster_name"),
						knownvalue.StringExact("test-cluster1"),
					),
					statecheck.ExpectKnownValue(
						"gsolaceclustermgr_broker.test2",
						tfjsonpath.New("custom_router_name"),
						knownvalue.StringExact("test-router1"),
					),
					statecheck.ExpectKnownValue(
						"gsolaceclustermgr_broker.test2",
						tfjsonpath.New("event_broker_version"),
						knownvalue.StringExact("1.0.0"),
					),
					// Verify Computed attributes
					statecheck.ExpectKnownValue(
						"gsolaceclustermgr_broker.test2",
						tfjsonpath.New("status"),
						knownvalue.StringExact("COMPLETED"),
					),
					// Verify dynamic values have any value set in the state.
					statecheck.ExpectKnownValue(
						"gsolaceclustermgr_broker.test2",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"gsolaceclustermgr_broker.test2",
						tfjsonpath.New("last_updated"),
						knownvalue.NotNull(),
					),
				},
			},
			// Update and Read testing
			{
				Config: testConfig("test2", "ocs-prov-test-changed", false),
				//ExpectNonEmptyPlan: true,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"gsolaceclustermgr_broker.test2",
						tfjsonpath.New("name"),
						knownvalue.StringExact("ocs-prov-test-changed"),
					), statecheck.ExpectKnownValue(
						"gsolaceclustermgr_broker.test2",
						tfjsonpath.New("last_updated"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"gsolaceclustermgr_broker.test2",
						tfjsonpath.New("status"),
						knownvalue.StringExact("PENDING"),
					),
				},
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testConfig(rname string, name string, allValues bool) string {
	var optionals = ""
	if allValues {
		optionals = `msg_vpn_name    = "ocs-msgvpn"
					cluster_name    = "gwc-aks-ocs"
					custom_router_name = "ocs-router"
					event_broker_version = "1.2.3"`
	}
	return providerConfig + `
	resource "gsolaceclustermgr_broker" "` + rname + `" {
		serviceclass_id = "ENTERPRISE_250_STANDALONE"
		name            = "` + name + `"
		datacenter_id   = "aks-germanywestcentral"
	    ` + optionals + `		
	}
	`
}

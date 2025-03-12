package provider

import (
	"fmt"
	"os"
	"regexp"
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
	svr = fakeserver.NewFakeServer(port, apiServerObjects, true, os.Getenv("FAKE_SERVER_DEBUG") != "", 0)
}

func stopFakeServer() {
	if svr != nil {
		fmt.Printf("Shutting down fake server ...\n")
		svr.Shutdown()
	}
}

func TestAccBrokerResource(t *testing.T) {
	if os.Getenv("FAKE_SERVER_EXT") == "" {
		startFakeServer()
		defer stopFakeServer()
	}
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testResourceConfig("test", "ocs-prov-test", true),
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
					statecheck.ExpectKnownValue(
						"gsolaceclustermgr_broker.test",
						tfjsonpath.New("max_spool_usage"),
						knownvalue.Int32Exact(23),
					),
					// Verify Computed attributes
					statecheck.ExpectKnownValue(
						"gsolaceclustermgr_broker.test",
						tfjsonpath.New("status"),
						knownvalue.StringExact("COMPLETED"),
					),
					statecheck.ExpectKnownValue(
						"gsolaceclustermgr_broker.test",
						tfjsonpath.New("hostnames"),
						knownvalue.ListExact([]knownvalue.Check{
							knownvalue.StringExact("test-host1"),
							knownvalue.StringExact("test-host2"),
						}),
					),
					statecheck.ExpectKnownValue(
						"gsolaceclustermgr_broker.test",
						tfjsonpath.New("service_endpoint_id"),
						knownvalue.StringExact("test-endpoint"),
					),
					statecheck.ExpectKnownValue(
						"gsolaceclustermgr_broker.test",
						tfjsonpath.New("missioncontrol_username"),
						knownvalue.StringExact("mc-user"),
					),
					statecheck.ExpectKnownValue(
						"gsolaceclustermgr_broker.test",
						tfjsonpath.New("missioncontrol_password"),
						knownvalue.StringExact("mc-passwd"),
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
				Config: testResourceConfig("test2", "ocs-prov-test2", false),
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
					statecheck.ExpectKnownValue(
						"gsolaceclustermgr_broker.test2",
						tfjsonpath.New("max_spool_usage"),
						knownvalue.Int32Exact(20),
					),
					statecheck.ExpectKnownValue(
						"gsolaceclustermgr_broker.test2",
						tfjsonpath.New("hostnames"),
						knownvalue.ListExact([]knownvalue.Check{
							knownvalue.StringExact("test-host1"),
							knownvalue.StringExact("test-host2"),
						}),
					),
					statecheck.ExpectKnownValue(
						"gsolaceclustermgr_broker.test2",
						tfjsonpath.New("service_endpoint_id"),
						knownvalue.StringExact("test-endpoint"),
					),
					statecheck.ExpectKnownValue(
						"gsolaceclustermgr_broker.test2",
						tfjsonpath.New("missioncontrol_username"),
						knownvalue.StringExact("mc-user"),
					),
					statecheck.ExpectKnownValue(
						"gsolaceclustermgr_broker.test2",
						tfjsonpath.New("missioncontrol_password"),
						knownvalue.StringExact("mc-passwd"),
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
			// Update and Read testing   (think about this again)
			{
				Config: testResourceConfig("test2", "ocs-prov-test-changed", false),
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
						knownvalue.StringExact("COMPLETED"),
					),
				},
			},
			// Delete testing automatically occurs in TestCase
		},
	})

}

func TestAccBrokerDataSource(t *testing.T) {
	if os.Getenv("EXT_SERVER") == "" {
		startFakeServer()
		defer stopFakeServer()
	}
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read fail testing
			{
				Config:      testDataSourceConfig("test3dsFail", "NotExisting1"),
				ExpectError: regexp.MustCompile("Error getting broker service info"),
			},
			// Read testing
			{
				PreConfig: func() {
					if svr != nil {
						svr.SetBaseSid(1234)
					}
				},
				Config: testResoureAndDataSourceConfig("test3ds"),
				ConfigStateChecks: []statecheck.StateCheck{
					// verify success:  attributes
					statecheck.ExpectKnownValue(
						"data.gsolaceclustermgr_broker.test3ds",
						tfjsonpath.New("id"),
						knownvalue.StringExact("1234"),
					),
					statecheck.ExpectKnownValue(
						"data.gsolaceclustermgr_broker.test3ds",
						tfjsonpath.New("msg_vpn_name"),
						knownvalue.StringExact("ocs-msgvpn"),
					),
					statecheck.ExpectKnownValue(
						"data.gsolaceclustermgr_broker.test3ds",
						tfjsonpath.New("cluster_name"),
						knownvalue.StringExact("gwc-aks-ocs"),
					),
					statecheck.ExpectKnownValue(
						"data.gsolaceclustermgr_broker.test3ds",
						tfjsonpath.New("custom_router_name"),
						knownvalue.StringExact("ocs-router"),
					),
					statecheck.ExpectKnownValue(
						"data.gsolaceclustermgr_broker.test3ds",
						tfjsonpath.New("event_broker_version"),
						knownvalue.StringExact("1.2.3"),
					),
					statecheck.ExpectKnownValue(
						"data.gsolaceclustermgr_broker.test3ds",
						tfjsonpath.New("max_spool_usage"),
						knownvalue.Int32Exact(23),
					),
					statecheck.ExpectKnownValue(
						"data.gsolaceclustermgr_broker.test3ds",
						tfjsonpath.New("status"),
						knownvalue.StringExact("COMPLETED"),
					),
					statecheck.ExpectKnownValue(
						"data.gsolaceclustermgr_broker.test3ds",
						tfjsonpath.New("last_updated"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"data.gsolaceclustermgr_broker.test3ds",
						tfjsonpath.New("hostnames"),
						knownvalue.ListExact([]knownvalue.Check{
							knownvalue.StringExact("test-host1"),
							knownvalue.StringExact("test-host2"),
						}),
					),
					statecheck.ExpectKnownValue(
						"data.gsolaceclustermgr_broker.test3ds",
						tfjsonpath.New("service_endpoint_id"),
						knownvalue.StringExact("test-endpoint"),
					),
					statecheck.ExpectKnownValue(
						"data.gsolaceclustermgr_broker.test3ds",
						tfjsonpath.New("missioncontrol_username"),
						knownvalue.StringExact("mc-user"),
					),
					statecheck.ExpectKnownValue(
						"data.gsolaceclustermgr_broker.test3ds",
						tfjsonpath.New("missioncontrol_password"),
						knownvalue.StringExact("mc-passwd"),
					),
				},
			},
		},
	})
}

func testResourceConfig(rname string, name string, allValues bool) string {
	var optionals = ""
	if allValues {
		optionals = `msg_vpn_name    = "ocs-msgvpn"
					cluster_name    = "gwc-aks-ocs"
					custom_router_name = "ocs-router"
					event_broker_version = "1.2.3"
					max_spool_usage = 23`
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

func testDataSourceConfig(rname string, id string) string {
	return providerConfig + `
	data "gsolaceclustermgr_broker" "` + rname + `" {
		id            = "` + id + `"
	}
	`
}

func testResoureAndDataSourceConfig(rname string) string {
	return testResourceConfig(rname, "foo", true) + `
	data "gsolaceclustermgr_broker" "` + rname + `" {
		id            = gsolaceclustermgr_broker.` + rname + `.id
	}
	`
}

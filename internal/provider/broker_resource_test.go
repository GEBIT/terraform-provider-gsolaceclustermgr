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
			// validation errors
			{
				Config:      testResourceConfigAll("test", "ocs-prov-test", "invalid-router!", 23),
				ExpectError: regexp.MustCompile("Invalid Attribute Value Length"),
			},
			{
				Config:      testResourceConfigAll("test", "ocs-prov-test", "ocsrouter", 1),
				ExpectError: regexp.MustCompile("Invalid Attribute Value"),
			},
			{
				Config: testResourceConfigAll("test", "ocs-prov-test", "ocsrouter", 23),
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
						knownvalue.StringExact("ocsrouter"),
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
					statecheck.ExpectKnownValue(
						"gsolaceclustermgr_broker.test",
						tfjsonpath.New("admin_username"),
						knownvalue.StringExact("ma-user"),
					),
					statecheck.ExpectKnownValue(
						"gsolaceclustermgr_broker.test",
						tfjsonpath.New("admin_password"),
						knownvalue.StringExact("ma-passwd"),
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
						knownvalue.StringExact(""),
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
				Config: testResourceConfig("test2", "ocs-prov-test2"),
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
						knownvalue.StringExact("testrouter1"),
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
					statecheck.ExpectKnownValue(
						"gsolaceclustermgr_broker.test2",
						tfjsonpath.New("admin_username"),
						knownvalue.StringExact("ma-user"),
					),
					statecheck.ExpectKnownValue(
						"gsolaceclustermgr_broker.test2",
						tfjsonpath.New("admin_password"),
						knownvalue.StringExact("ma-passwd"),
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
				Config: testResourceConfig("test2", "ocs-prov-test-changed"),
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
				ExpectError: regexp.MustCompile("Could not find broker service for id \"NotExisting1\""),
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
						knownvalue.StringExact("ocsrouterprimarycn"),
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
					statecheck.ExpectKnownValue(
						"data.gsolaceclustermgr_broker.test3ds",
						tfjsonpath.New("admin_username"),
						knownvalue.StringExact("ma-user"),
					),
					statecheck.ExpectKnownValue(
						"data.gsolaceclustermgr_broker.test3ds",
						tfjsonpath.New("admin_password"),
						knownvalue.StringExact("ma-passwd"),
					),
				},
			},
		},
	})
}

func testResourceConfigAll(rname string, name string, routerName string, spoolSize int) string {

	optionals := `msg_vpn_name    = "ocs-msgvpn"
		cluster_name    = "gwc-aks-ocs"
		custom_router_name = "` + routerName + `"
		event_broker_version = "1.2.3"
		max_spool_usage = ` + fmt.Sprint(spoolSize)
	return providerConfig + `
	resource "gsolaceclustermgr_broker" "` + rname + `" {
		serviceclass_id = "ENTERPRISE_250_STANDALONE"
		name            = "` + name + `"
		datacenter_id   = "aks-germanywestcentral"
		` + optionals + `		
	}
	`
}
func testResourceConfig(rname string, name string) string {
	return providerConfig + `
	resource "gsolaceclustermgr_broker" "` + rname + `" {
		serviceclass_id = "ENTERPRISE_250_STANDALONE"
		name            = "` + name + `"
		datacenter_id   = "aks-germanywestcentral"
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
	return testResourceConfigAll(rname, "foo", "ocsrouter", 23) + `
	data "gsolaceclustermgr_broker" "` + rname + `" {
		id            = gsolaceclustermgr_broker.` + rname + `.id
	}
	`
}

package provider

import (
	"fmt"
	"terraform-provider-gsolaceclustermgr/internal/fakeserver"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

var svr *fakeserver.Fakeserver

func startFakeServer() {
	apiServerObjects := make(map[string]fakeserver.ServiceInfo)
	port := 8091
	fmt.Printf("Starting fake server on port %d...\n", port)
	svr = fakeserver.NewFakeServer(port, apiServerObjects, true, true)
}

func stopFakeServer() {
	if svr != nil {
		fmt.Printf("Shutting down fake server ...\n")
		svr.Shutdown()
	}
}

func TestAccBrokerResource(t *testing.T) {
	startFakeServer()
	defer stopFakeServer()
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
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify attributes
					resource.TestCheckResourceAttr("gsolaceclustermgr_broker.test", "name", "ocs-prov-test"),
					resource.TestCheckResourceAttr("gsolaceclustermgr_broker.test", "serviceclass_id", "ENTERPRISE_250_STANDALONE"),
					resource.TestCheckResourceAttr("gsolaceclustermgr_broker.test", "datacenter_id", "aks-germanywestcentral"),
					resource.TestCheckResourceAttr("gsolaceclustermgr_broker.test", "msg_vpn_name", "ocs-msgvpn-1"),
					resource.TestCheckResourceAttr("gsolaceclustermgr_broker.test", "cluster_name", "gwc-aks-cluster1"),

					// Verify Computed attributes
					resource.TestCheckResourceAttr("gsolaceclustermgr_broker.test", "status", "COMPLETED"),
					// Verify dynamic values have any value set in the state.
					resource.TestCheckResourceAttrSet("gsolaceclustermgr_broker.test", "id"),
					resource.TestCheckResourceAttrSet("gsolaceclustermgr_broker.test", "last_updated"),
				),
			},

			// Delete testing automatically occurs in TestCase
		},
	})
}

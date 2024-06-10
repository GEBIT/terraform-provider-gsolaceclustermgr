package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccBrokerResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig + `
resource "clustermanager_broker" "test" {
	serviceclass_id = "ENTERPRISE_250_STANDALONE"
	name            = "ocs-prov-test"
	datacenter_id   = "aks-germanywestcentral"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify attributes
					resource.TestCheckResourceAttr("clustermanager_broker.test", "name", "ocs-prov-test"),
					resource.TestCheckResourceAttr("clustermanager_broker.test", "serviceclass_id", "ENTERPRISE_250_STANDALONE"),
					resource.TestCheckResourceAttr("clustermanager_broker.test", "datacenter_id", "aks-germanywestcentralE"),

					// Verify Computed attributes
					resource.TestCheckResourceAttr("hashicups_order.test", "status", "pending"),
					// Verify dynamic values have any value set in the state.
					resource.TestCheckResourceAttrSet("hashicups_order.test", "id"),
					resource.TestCheckResourceAttrSet("hashicups_order.test", "last_updated"),
				),
			},

			// Delete testing automatically occurs in TestCase
		},
	})
}

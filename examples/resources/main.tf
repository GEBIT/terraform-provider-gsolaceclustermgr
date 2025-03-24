terraform {
  required_providers {
    gsolaceclustermgr = {
      source = "GEBIT/gsolaceclustermgr"
    }
  }
  required_version = ">= 1.1.0"
}

provider "gsolaceclustermgr" {
  ###### test against solace cloud 
  # bearer_token = "<aSolaceApiToken"
  # host = "https://api.solace.cloud"

  ###### test against fakeserver
  bearer_token              = "bt42"
  host                      = "http://localhost:8091"
  polling_interval_duration = "2s"

}

resource "gsolaceclustermgr_broker" "ocs-test" {
  serviceclass_id = "ENTERPRISE_250_STANDALONE"
  name            = "ocs-prov-test-1"
  datacenter_id   = "aks-germanywestcentral"

  # optional attributes

  msg_vpn_name = "ocs-msgvpn-1"
  #cluster_name    = "gwc-aks-cluster1"
  #custom_router_name = "ocsrouter1"
  #event_broker_version = "10.8.1.152-7"
  #max_spool_usage = 40
}

output "ocs-test_id" {
  value = gsolaceclustermgr_broker.ocs-test.id
}
output "ocs-test" {
  value     = gsolaceclustermgr_broker.ocs-test
  sensitive = true
}

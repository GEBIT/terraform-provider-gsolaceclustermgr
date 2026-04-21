terraform {
  required_providers {
    gsolaceclustermgr = {
      source = "GEBIT/gsolaceclustermgr"
    }
  }
  required_version = ">= 1.1.0"
}

variable "bearer_token" {
  type        = string
  sensitive   = true
  description = "Solace Cloud API bearer token. Set via TF_VAR_bearer_token."
}

provider "gsolaceclustermgr" {
  ###### test against solace cloud 
  bearer_token = var.bearer_token
  host         = "https://api.solace.cloud"

  ###### test against fakeserver
  # bearer_token              = "bt42"
  # host                      = "http://localhost:8091"
  polling_interval_duration = "2s"

}

resource "gsolaceclustermgr_broker" "ocs-test" {
  serviceclass_id = "ENTERPRISE_250_STANDALONE"
  name            = "ocs-prov-test-1"
  datacenter_id   = "aks-germanywestcentral"

  # optional attributes

  msg_vpn_name       = "ocs-msgvpn-1"
  cluster_name       = "gebit-test"
  custom_router_name = "ocsrouter1"
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

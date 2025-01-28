terraform {
  required_providers {
    gsolaceclustermgr = {
      source = "gebit.de/tf/gsolaceclustermgr"
    }
  }
  required_version = ">= 1.1.0"
}

provider "gsolaceclustermgr" {
  // insert a valid token here
  bearer_token = "<solaceAPIToken>"

  host = "https://api.solace.cloud"
}

resource "gsolaceclustermgr_broker" "ocs-test" {
  count           = 1
  serviceclass_id = "ENTERPRISE_250_STANDALONE"
  name            = "ocs-prov-test"
  datacenter_id   = "aks-germanywestcentral"
  msg_vpn_name    = "ocs-msgvpn-1"
  cluster_name    = "gwc-aks-cluster1"
}


output "broker_ocs-test" {
  value = gsolaceclustermgr_broker.ocs-test
}

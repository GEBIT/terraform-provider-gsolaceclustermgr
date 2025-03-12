terraform {
  required_providers {
    gsolaceclustermgr = {
      source = "gebit.de/tf/gsolaceclustermgr"
    }
  }
  required_version = ">= 1.1.0"
}

provider "gsolaceclustermgr" {
  ####### test against solace cloud 
  ## bearer_token = "<aSolaceApiToken>"
  ## host = "https://api.solace.cloud"

  ####### test against fakeserver
  bearer_token              = "bt42"
  host                      = "http://localhost:8091"
  polling_interval_duration = "2s"

}

data "gsolaceclustermgr_broker" "ocs-test" {
  id = "25edf6cb-7b85-4a42-9be3-86736d803242"
}


output "ocs-test_id" {
  value = data.gsolaceclustermgr_broker.ocs-test.id
}
output "ocs-test" {
  value     = data.gsolaceclustermgr_broker.ocs-test
  sensitive = true
}

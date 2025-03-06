terraform {
  required_providers {
    gsolaceclustermgr = {
      source = "gebit.de/tf/gsolaceclustermgr"
    }
  }
  required_version = ">= 1.1.0"
}

provider "gsolaceclustermgr" {
  // --- test against solace cloud 
  // bearer_token = "<aSolaceApiToken"
  // host = "https://api.solace.cloud"

  // --- test against fakeserver
  bearer_token              = "bt42"
  host                      = "http://localhost:8091"
  polling_interval_duration = "2s"

}

data "gsolaceclustermgr_broker" "ocs-test" {
  id = "f4d0e212-a6e0-40a3-8d3e-7ad897228a75"
}


output "ocs-test_id" {
  value = data.gsolaceclustermgr_broker.ocs-test.id
}
output "ocs-test_secret" {
  value = data.gsolaceclustermgr_broker.ocs-test.client_secret
  sensitive = true
}

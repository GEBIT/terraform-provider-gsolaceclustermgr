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
  bearer_token = "eyJhbGciOiJSUzI1NiIsImtpZCI6Im1hYXNfcHJvZF8yMDIwMDMyNiIsInR5cCI6IkpXVCJ9.eyJvcmciOiJnZWJpdCIsIm9yZ1R5cGUiOiJFTlRFUlBSSVNFIiwic3ViIjoiN2o4N2dpcHp6OHYiLCJwZXJtaXNzaW9ucyI6IkFBQUFBQUFBQUFBQVd3QUFZQUFBQUFBQUFBQUFBQUFBQUFBQUFnQUFBQUFBQUFBQUFBQkFBSmdBV0FBQWdBQUFBQUFBRUFJPSIsImFwaVRva2VuSWQiOiJlaDY5Y2NmYjBwYSIsImlzcyI6IlNvbGFjZSBDb3Jwb3JhdGlvbiIsImlhdCI6MTcxMzk0NTQ2MH0.XuW_eR--retS-H7-36vz10DRfHFu6tBOi5P2xtnJVY4mBd7eYUG3prHGEL8b4-afFe61aji343CxDCvBYc4HdkYvGO44NgtPhFQTotXefVNHo2E3tgcF1xdAQDgB1NB5YmGu6vWP9OdXiB6oJmy6lfnZVek34Ypw8G_wx3Se4vC86XQuQAv4AXgfqP3NtGUBSqYyURt6JVAgvk1muRNjuuhhsRMyGhn7x0EKVBtpL7h7atRCc0zKn4K83fM7Olf9GBkjgHAyJCC7LsC2agBGHPKy9gIw8W4tjJr6MJOhWnB3yrvImikUoI5cgS1l_EICjoiSAeGEBcYgDXidypuQLg"

  host = "https://api.solace.cloud"
}

resource "gsolaceclustermgr_broker" "ocs-test" {
  count           = 1
  serviceclass_id = "ENTERPRISE_250_STANDALONE"
  name            = "ocs-prov-testX"
  datacenter_id   = "aks-germanywestcentral"
  msg_vpn_name    = "ocs-msgvpn-1"
  cluster_name    = "gwc-aks-cluster1"
}


output "broker_ocs-test" {
  value = gsolaceclustermgr_broker.ocs-test
}

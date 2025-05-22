terraform {
  required_providers {
    ios = {
      source = "hashicorp.com/edu/ios"
    }
  }
}


provider "ios" {
  host     = "192.168.100.200"
  username = "admin"
  password = "MyStrongPassword"
}

resource "ios_vlan" "test" {
  name = "test"
  id   = 1104
}

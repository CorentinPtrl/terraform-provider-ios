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

data "ios_vlans" "test" {
}

data "ios_vlan" "test" {
  id = 1104
}

data "ios_static_routes" "test" {
}

resource "ios_static_route" "test" {
  prefix = "192.168.51.0"
  mask   = "255.255.255.0"
  next_hop = "1.1.1.1"
}

output "vlans" {
  value = data.ios_vlans.test
}

output "vlan" {
  value = data.ios_vlan.test
}

output "static_routes" {
  value = data.ios_static_routes.test
}
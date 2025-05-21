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

data "ios_show_cdp_neighbors" "test" {}

data "ios_show_vlan" "test" {}

data "ios_show_users" "test" {}

data "ios_show_interfaces" "test" {}

data "ios_show_ip_route" "test" {}

output "cdp" {
  value = data.ios_show_cdp_neighbors.test
}

output "vlan" {
  value = data.ios_show_vlan.test
}

output "users" {
  value = data.ios_show_users.test
}

output "interfaces" {
  value = data.ios_show_interfaces.test
}

output "ip_route" {
  value = data.ios_show_ip_route.test
}
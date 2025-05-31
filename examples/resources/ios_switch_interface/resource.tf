resource "ios_switch_interface" "access" {
  id       = "GigabitEthernet0/0"
  shutdown = false
  access = {
    access_vlan = 2
  }
  trunk = {
    encapsulation = "dot1q"
    allowed_vlans = [10, 20, 30, 40]
  }
  spanning_tree = {
    portfast   = "edge"
    bpdu_guard = true
  }
  description = "Access interface for VLAN 2 with trunking and spanning tree settings"
}

resource "ios_switch_interface" "trunk" {
  id       = "GigabitEthernet0/1"
  shutdown = false
  trunk = {
    encapsulation = "dot1q"
    allowed_vlans = [100, 200, 300]
  }
  description = "Trunk interface with specific VLANs allowed"
}
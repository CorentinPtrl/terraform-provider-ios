resource "ios_ethernet_interface" "example" {
  id       = "GigabitEthernet0/0"
  shutdown = false
  ips = [
    {
      ip = "192.198.101.1/24"
  }]
  description = "Test description of GigabitEthernet0/0"
}
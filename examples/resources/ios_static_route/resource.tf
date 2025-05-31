resource "ios_static_route" "example" {
  prefix   = "192.168.51.0"
  mask     = "255.255.255.0"
  next_hop = "192.168.20.1"
}

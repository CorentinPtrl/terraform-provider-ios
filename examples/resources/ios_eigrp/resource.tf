resource "ios_eigrp" "example" {
  as_number = 5
  networks = [
    "192.168.80.0/24",
  ]
}
terraform {
  required_providers {
    ios = {
      source = "hashicorp.com/edu/ios"
    }
  }
}

provider "ios" {}

data "ios_example" "example" {}

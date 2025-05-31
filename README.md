# Terraform Provider for Cisco IOS

This project is a Terraform provider for managing Cisco IOS devices. It allows users to configure network resources such as VLANs, interfaces, static routes, and EIGRP processes using Terraform.

## Features

- **VLAN Management**: Create, update, read, and delete VLANs.
- **Interface Configuration**: Manage Ethernet and switch interfaces, including trunk and access modes.
- **Static Routes**: Configure static routes for network traffic.
- **EIGRP**: Manage EIGRP processes and associated networks.
- **Data Sources**: Retrieve information about VLANs, interfaces, and static routes.


## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.24

## Building The Provider

1. Clone the repository
2. Enter the repository directory
3. Build the provider using the Go `install` command:

```shell
go install
```

## Developing the Provider

If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed on your machine (see [Requirements](#requirements) above).

To compile the provider, run `go install`. This will build the provider and put the provider binary in the `$GOPATH/bin` directory.

To generate or update documentation, run `make generate`.

In order to run the full suite of Acceptance tests, run `make testacc`.

*Note:* Acceptance tests create real resources.

```shell
make testacc
```

## Contributing
Contributions are welcome! Please submit issues or pull requests to improve the provider.

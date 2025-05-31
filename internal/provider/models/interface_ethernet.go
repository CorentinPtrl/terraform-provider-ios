// Copyright (c) Corentin Pitrel
// SPDX-License-Identifier: MIT

package models

import (
	"context"
	"fmt"
	"github.com/CorentinPtrl/cgnet"
	"github.com/CorentinPtrl/cisconf"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"net"
	"terraform-provider-ios/internal/utils"
)

type InterfaceEthernetModel struct {
	Ips             types.List `tfsdk:"ips"`
	HelperAddresses types.List `tfsdk:"helper_addresses"`
	InterfaceModel
}

type IpInterfaceModel struct {
	Ip types.String `tfsdk:"ip"`
}

func (ip IpInterfaceModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"ip": types.StringType,
	}
}

func (ip IpInterfaceModel) AttributeValues() map[string]attr.Value {
	return map[string]attr.Value{
		"ip": ip.Ip,
	}
}

func InterfaceEthernetFromCisconf(ctx context.Context, iface *cisconf.CiscoInterface) (InterfaceEthernetModel, error) {
	ipsModel := make([]IpInterfaceModel, len(iface.Ips))
	for i, ip := range iface.Ips {
		cidr, err := utils.SubnetMaskToCIDR(ip.Subnet)
		if err != nil {
			return InterfaceEthernetModel{}, fmt.Errorf("failed to convert subnet %s to CIDR: %v", ip.Subnet, err)
		}
		ipsModel[i] = IpInterfaceModel{
			Ip: types.StringValue(fmt.Sprintf("%s/%d", ip.Ip, cidr)),
		}
	}
	ips, err := types.ListValueFrom(ctx, types.ObjectType{}.WithAttributeTypes(IpInterfaceModel{}.AttributeTypes()), ipsModel)
	if err != nil {
		return InterfaceEthernetModel{}, fmt.Errorf("failed to convert IP list to ListValue: %v", err)
	}
	helperAdresses, diags := types.ListValueFrom(ctx, types.StringType, iface.IPHelperAddresses)
	if diags.HasError() {
		return InterfaceEthernetModel{}, fmt.Errorf("failed to convert helper addresses to ListValue: %v", diags)
	}
	return InterfaceEthernetModel{
		InterfaceModel: InterfaceModel{
			ID:          types.StringValue(iface.Parent.Identifier),
			Description: types.StringValue(iface.Description),
			Shutdown:    types.BoolValue(iface.Shutdown),
		},
		Ips:             ips,
		HelperAddresses: helperAdresses,
	}, nil
}

func InterfaceEthernetToCisconf(ctx context.Context, iface InterfaceEthernetModel) (*cisconf.CiscoInterface, error) {
	cisIface := &cisconf.CiscoInterface{
		Parent: cisconf.CiscoInterfaceParent{
			Identifier: iface.ID.ValueString(),
		},
		Description: iface.Description.ValueString(),
		Shutdown:    iface.Shutdown.ValueBool(),
	}
	cisIface.Switchport = false
	var ips []IpInterfaceModel
	err := iface.Ips.ElementsAs(ctx, &ips, false)
	if err != nil {
		return nil, fmt.Errorf("failed to convert IP list to slice: %v", err)
	}
	for _, ip := range ips {
		if ip.Ip.IsNull() || ip.Ip.ValueString() == "" {
			return nil, fmt.Errorf("IP address is null or empty")
		}
		host, ipNet, err := net.ParseCIDR(ip.Ip.ValueString())
		if err != nil {
			return nil, fmt.Errorf("failed to parse CIDR %s: %v", ip.Ip.ValueString(), err)
		}
		mask := net.IP(ipNet.Mask).String()
		cisIface.Ips = append(cisIface.Ips, cisconf.Ip{
			Ip:     host.String(),
			Subnet: mask,
		})
	}
	var helperAddresses []string
	err = iface.HelperAddresses.ElementsAs(ctx, &helperAddresses, false)
	if err != nil {
		return nil, fmt.Errorf("failed to convert helper addresses to slice: %v", err)
	}
	cisIface.IPHelperAddresses = helperAddresses
	return cisIface, nil
}

func GetEthernetInterfaces(ctx context.Context, device *cgnet.Device) ([]InterfaceEthernetModel, error) {
	config, err := device.Exec("sh running-config")
	if err != nil {
		return nil, fmt.Errorf("failed to execute running config: %w", err)
	}
	var runningConfig cisconf.Config
	err = cisconf.Unmarshal(config, &runningConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal running config: %w", err)
	}
	result := []InterfaceEthernetModel{}
	for _, inter := range runningConfig.Interfaces {
		var interfaceEthernet InterfaceEthernetModel
		interfaceEthernet, err = InterfaceEthernetFromCisconf(ctx, &inter)
		if err != nil {
			return nil, fmt.Errorf("failed to convert interface: %w", err)
		}
		result = append(result, interfaceEthernet)
	}
	return result, nil
}

func GetEthernetInterface(ctx context.Context, device *cgnet.Device, interfaceID string) (InterfaceEthernetModel, error) {
	interfaces, err := GetEthernetInterfaces(ctx, device)
	if err != nil {
		return InterfaceEthernetModel{}, fmt.Errorf("failed to get Ethernet interfaces: %w", err)
	}

	for _, inter := range interfaces {
		if inter.ID.ValueString() == interfaceID {
			return inter, nil
		}
	}
	return InterfaceEthernetModel{}, fmt.Errorf("interface %s not found", interfaceID)
}

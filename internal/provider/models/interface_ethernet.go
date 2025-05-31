// Copyright (c) Corentin Pitrel
// SPDX-License-Identifier: MIT

package models

import (
	"context"
	"fmt"
	"github.com/CorentinPtrl/cisconf"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
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

func InterfaceEthernetFromCisconf(ctx context.Context, iface *cisconf.CiscoInterface) InterfaceEthernetModel {
	ipsModel := make([]IpInterfaceModel, len(iface.Ips))
	for i, ip := range iface.Ips {
		cidr, err := utils.SubnetMaskToCIDR(ip.Subnet)
		if err != nil {
			tflog.Error(ctx, fmt.Sprintf("Failed to convert subnet %s to CIDR: %v", ip.Subnet, err))
			return InterfaceEthernetModel{}
		}
		ipsModel[i] = IpInterfaceModel{
			Ip: types.StringValue(fmt.Sprintf("%s/%d", ip.Ip, cidr)),
		}
	}
	ips, err := types.ListValueFrom(ctx, types.ObjectType{}.WithAttributeTypes(IpInterfaceModel{}.AttributeTypes()), ipsModel)
	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("Failed to convert IP list to ListValue: %v", err))
		return InterfaceEthernetModel{}
	}
	helperAdresses, diags := types.ListValueFrom(ctx, types.StringType, iface.IPHelperAddresses)
	if diags.HasError() {
		tflog.Error(ctx, fmt.Sprintf("Failed to convert helper addresses to ListValue: %v", diags))
		return InterfaceEthernetModel{}
	}
	return InterfaceEthernetModel{
		InterfaceModel: InterfaceModel{
			ID:          types.StringValue(iface.Parent.Identifier),
			Description: types.StringValue(iface.Description),
			Shutdown:    types.BoolValue(iface.Shutdown),
		},
		Ips:             ips,
		HelperAddresses: helperAdresses,
	}
}

func InterfaceEthernetToCisconf(ctx context.Context, iface InterfaceEthernetModel) *cisconf.CiscoInterface {
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
		tflog.Error(ctx, fmt.Sprintf("Failed to convert IP list to slice: %v", err))
		return nil
	}
	for _, ip := range ips {
		if ip.Ip.IsNull() || ip.Ip.ValueString() == "" {
			tflog.Error(ctx, "IP address is null or empty")
			return nil
		}
		host, ipNet, err := net.ParseCIDR(ip.Ip.ValueString())
		if err != nil {
			tflog.Info(ctx, fmt.Sprintf("Failed to parse CIDR %s: %v", ip.Ip.ValueString(), err))
			return nil
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
		tflog.Error(ctx, fmt.Sprintf("Failed to convert helper addresses to slice: %v", err))
		return nil
	}
	cisIface.IPHelperAddresses = helperAddresses
	return cisIface
}

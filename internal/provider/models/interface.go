package models

import (
	"context"
	"github.com/CorentinPtrl/cisconf"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

type InterfacesDataSourceModel struct {
	Interfaces []InterfaceModel `tfsdk:"interfaces"`
}

type InterfaceModel struct {
	ID          types.String `tfsdk:"id"`
	Switchport  types.String `tfsdk:"switchport"`
	Ips         types.List   `tfsdk:"ips"`
	Description types.String `tfsdk:"description"`
	Shutdown    types.Bool   `tfsdk:"shutdown"`
}

type IpInterfaceModel struct {
	Ip   types.String `tfsdk:"ip"`
	Mask types.String `tfsdk:"mask"`
}

func (ip IpInterfaceModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"ip":   types.StringType,
		"mask": types.StringType,
	}
}

func InterfaceFromCisconf(ctx context.Context, iface *cisconf.CiscoInterface) InterfaceModel {
	ips := []types.Object{}
	for _, ip := range iface.Ips {
		value := IpInterfaceModel{
			Ip:   types.StringValue(ip.Ip),
			Mask: types.StringValue(ip.Subnet),
		}
		obj, _ := types.ObjectValueFrom(ctx, value.AttributeTypes(), value)
		ips = append(ips, obj)

	}
	var switchport types.String
	if iface.Switchport {
		if iface.Trunk {
			switchport = types.StringValue("trunk")
		} else {
			switchport = types.StringValue("access")
		}
	} else {
		switchport = types.StringNull()
	}

	switchport = types.StringNull()

	listValue, _ := types.ListValueFrom(ctx, types.ObjectType{}.WithAttributeTypes(IpInterfaceModel{}.AttributeTypes()), ips)

	return InterfaceModel{
		ID:          types.StringValue(iface.Parent.Identifier),
		Switchport:  switchport,
		Ips:         listValue,
		Description: types.StringValue(iface.Description),
		Shutdown:    types.BoolValue(iface.Shutdown),
	}
}

func InterfaceToCisconf(ctx context.Context, iface InterfaceModel) *cisconf.CiscoInterface {
	cisIface := &cisconf.CiscoInterface{
		Parent: cisconf.CiscoInterfaceParent{
			Identifier: iface.ID.ValueString(),
		},
		Description: iface.Description.ValueString(),
		Shutdown:    iface.Shutdown.ValueBool(),
	}

	if iface.Switchport.IsNull() {
		cisIface.Switchport = false
	} else if iface.Switchport.Equal(types.StringValue("trunk")) {
		cisIface.Switchport = true
		cisIface.Trunk = true
	} else {
		cisIface.Switchport = true
	}
	cisIface.Switchport = false
	ips := make([]types.Object, 0, len(iface.Ips.Elements()))
	diags := iface.Ips.ElementsAs(ctx, &ips, false)
	if diags.HasError() {
		return nil
	}
	for _, ip := range ips {
		var cisIp IpInterfaceModel
		ip.As(ctx, &cisIp, basetypes.ObjectAsOptions{})
		cisIface.Ips = append(cisIface.Ips, cisconf.Ip{
			Ip:     cisIp.Ip.ValueString(),
			Subnet: cisIp.Mask.ValueString(),
		})
	}
	return cisIface
}

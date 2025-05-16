package models

import (
	"context"
	"github.com/CorentinPtrl/cisconf"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type InterfacesSwitchesDataSourceModel struct {
	Interfaces []InterfaceSwitchModel `tfsdk:"interfaces"`
}

type InterfaceSwitchModel struct {
	Switchport            types.String `tfsdk:"switchport"`
	AccessVlan            types.Int32  `tfsdk:"access_vlan"`
	Encapsulation         types.String `tfsdk:"encapsulation"`
	AllowedVlans          types.List   `tfsdk:"allowed_vlans"`
	SpanningTreePortfast  types.String `tfsdk:"spanning_tree_portfast"`
	SpanningTreeBpduGuard types.Bool   `tfsdk:"spanning_tree_bpdu_guard"`
	InterfaceModel
}

type InterfaceEthernetModel struct {
	Ips types.List `tfsdk:"ips"`
	InterfaceModel
}

type InterfaceModel struct {
	ID          types.String `tfsdk:"id"`
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

func InterfaceSwitchFromCisconf(ctx context.Context, iface *cisconf.CiscoInterface) InterfaceSwitchModel {
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

	allowedVlans := types.ListNull(types.Int32Type)
	if iface.Trunk && iface.TrunkAllowedVlan != nil {
		var err diag.Diagnostics
		allowedVlans, err = types.ListValueFrom(ctx, types.Int32Type, iface.TrunkAllowedVlan)
		if err != nil {
			tflog.Error(ctx, "Failed to convert trunk allowed VLANs to list")
			return InterfaceSwitchModel{}
		}
	}

	return InterfaceSwitchModel{
		Switchport:            switchport,
		AccessVlan:            types.Int32Value(int32(iface.AccessVlan)),
		Encapsulation:         types.StringValue(iface.Encapsulation),
		AllowedVlans:          allowedVlans,
		SpanningTreePortfast:  types.StringValue(iface.STPPortFast),
		SpanningTreeBpduGuard: types.BoolValue(iface.STPBpduGuard == "enable"),
		InterfaceModel: InterfaceModel{
			ID:          types.StringValue(iface.Parent.Identifier),
			Description: types.StringValue(iface.Description),
			Shutdown:    types.BoolValue(iface.Shutdown),
		},
	}
}

func InterfaceSwitchToCisconf(ctx context.Context, iface InterfaceSwitchModel) *cisconf.CiscoInterface {
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
		cisIface.Access = false
		cisIface.Trunk = true
		cisIface.Encapsulation = iface.Encapsulation.ValueString()
		allowedVlans := make([]types.Int32, 0, len(iface.AllowedVlans.Elements()))
		diags := iface.AllowedVlans.ElementsAs(ctx, &allowedVlans, false)
		if diags.HasError() {
			return nil
		}
		cisIface.TrunkAllowedVlan = []int{}
		for _, vlan := range allowedVlans {
			cisIface.TrunkAllowedVlan = append(cisIface.TrunkAllowedVlan, int(vlan.ValueInt32()))
		}
	} else {
		cisIface.Switchport = true
		cisIface.Trunk = false
		cisIface.Access = true
		cisIface.AccessVlan = int(iface.AccessVlan.ValueInt32())
		cisIface.STPPortFast = iface.SpanningTreePortfast.ValueString()
		if iface.SpanningTreeBpduGuard.ValueBool() {
			cisIface.STPBpduGuard = "enable"
		} else {
			cisIface.STPBpduGuard = "disable"
		}
	}
	return cisIface
}

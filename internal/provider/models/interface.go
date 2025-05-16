package models

import (
	"context"
	"github.com/CorentinPtrl/cisconf"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type InterfacesSwitchesDataSourceModel struct {
	Interfaces []InterfaceSwitchModel `tfsdk:"interfaces"`
}

type InterfaceSwitchModel struct {
	Switchport   types.String          `tfsdk:"switchport"`
	Access       basetypes.ObjectValue `tfsdk:"access"`
	Trunk        basetypes.ObjectValue `tfsdk:"trunk"`
	SpanningTree basetypes.ObjectValue `tfsdk:"spanning_tree"`
	InterfaceModel
}

type Access struct {
	AccessVlan types.Int32 `tfsdk:"access_vlan"`
}

type Trunk struct {
	Encapsulation types.String `tfsdk:"encapsulation"`
	AllowedVlans  types.List   `tfsdk:"allowed_vlans"`
}

type SpanningTree struct {
	Portfast  types.String `tfsdk:"portfast"`
	BpduGuard types.Bool   `tfsdk:"bpdu_guard"`
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

func SpanningTreeFromObjectValue(ctx context.Context, obj basetypes.ObjectValue) (SpanningTree, diag.Diagnostics) {
	var st SpanningTree
	diags := obj.As(ctx, &st, basetypes.ObjectAsOptions{})
	if diags.HasError() {
		tflog.Error(ctx, "Failed to convert ObjectValue to SpanningTree")
		return SpanningTree{}, diags
	}
	return st, nil
}

func (st SpanningTree) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"portfast":   types.StringType,
		"bpdu_guard": types.BoolType,
	}
}

func (st SpanningTree) AttributeValues() map[string]attr.Value {
	return map[string]attr.Value{
		"portfast":   st.Portfast,
		"bpdu_guard": st.BpduGuard,
	}
}

func AccessFromObjectValue(ctx context.Context, obj basetypes.ObjectValue) (Access, diag.Diagnostics) {
	var access Access
	diags := obj.As(ctx, &access, basetypes.ObjectAsOptions{})
	if diags.HasError() {
		tflog.Error(ctx, "Failed to convert ObjectValue to Access")
		return Access{}, diags
	}
	return access, nil
}

func (access Access) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"access_vlan": types.Int32Type,
	}
}

func (access Access) AttributeValues() map[string]attr.Value {
	return map[string]attr.Value{
		"access_vlan": access.AccessVlan,
	}
}

func TrunkFromObjectValue(ctx context.Context, obj basetypes.ObjectValue) (Trunk, diag.Diagnostics) {
	var trunk Trunk
	diags := obj.As(ctx, &trunk, basetypes.ObjectAsOptions{})
	if diags.HasError() {
		tflog.Error(ctx, "Failed to convert ObjectValue to Trunk")
		return Trunk{}, diags
	}
	return trunk, nil
}

func (trunk Trunk) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"encapsulation": types.StringType,
		"allowed_vlans": types.ListType{ElemType: types.Int32Type},
	}
}

func (trunk Trunk) AttributeValues() map[string]attr.Value {
	return map[string]attr.Value{
		"encapsulation": trunk.Encapsulation,
		"allowed_vlans": trunk.AllowedVlans,
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
	st := SpanningTree{
		Portfast:  types.StringValue(iface.STPPortFast),
		BpduGuard: types.BoolValue(iface.STPBpduGuard == "enable"),
	}
	st_obj, diags := types.ObjectValue(st.AttributeTypes(), st.AttributeValues())
	if diags.HasError() {
		tflog.Error(ctx, "Failed to convert SpanningTree to object value")
		return InterfaceSwitchModel{}
	}

	access_obj := types.ObjectNull(Access{}.AttributeTypes())
	if iface.Access {
		access := Access{
			AccessVlan: types.Int32Value(int32(iface.AccessVlan)),
		}

		access_obj, diags = types.ObjectValue(access.AttributeTypes(), access.AttributeValues())
		if diags.HasError() {
			tflog.Error(ctx, "Failed to convert Access to object value")
			return InterfaceSwitchModel{}
		}
	}

	trunk_obj := types.ObjectNull(Trunk{}.AttributeTypes())
	if iface.Trunk {

		trunk := Trunk{
			Encapsulation: types.StringValue(iface.Encapsulation),
			AllowedVlans:  allowedVlans,
		}
		trunk_obj, diags = types.ObjectValue(trunk.AttributeTypes(), trunk.AttributeValues())
		if diags.HasError() {
			tflog.Error(ctx, "Failed to convert Trunk to object value")
			return InterfaceSwitchModel{}
		}
	}

	return InterfaceSwitchModel{
		Switchport:   switchport,
		Access:       access_obj,
		Trunk:        trunk_obj,
		SpanningTree: st_obj,
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

	if iface.Access.IsUnknown() && iface.Trunk.IsUnknown() {
		cisIface.Switchport = false
	} else if !(iface.Trunk.IsUnknown() || iface.Trunk.IsNull()) {
		cisIface.Switchport = true
		cisIface.Access = false
		cisIface.Trunk = true
		trunk, err := TrunkFromObjectValue(ctx, iface.Trunk)
		if err != nil {
			tflog.Error(ctx, "Failed to convert Trunk from ObjectValue")
			return nil
		}
		cisIface.Encapsulation = trunk.Encapsulation.ValueString()
		allowedVlans := make([]types.Int32, 0, len(trunk.AllowedVlans.Elements()))
		diags := trunk.AllowedVlans.ElementsAs(ctx, &allowedVlans, false)
		if diags.HasError() {
			return nil
		}
		cisIface.TrunkAllowedVlan = []int{}
		for _, vlan := range allowedVlans {
			cisIface.TrunkAllowedVlan = append(cisIface.TrunkAllowedVlan, int(vlan.ValueInt32()))
		}
	} else if !(iface.Access.IsUnknown() || iface.Access.IsNull()) {
		cisIface.Switchport = true
		cisIface.Trunk = false
		cisIface.Access = true
		access, err := AccessFromObjectValue(ctx, iface.Access)
		if err != nil {
			tflog.Error(ctx, "Failed to convert Access from ObjectValue")
			return nil
		}
		cisIface.AccessVlan = int(access.AccessVlan.ValueInt32())
	}
	if iface.SpanningTree.IsUnknown() {
		st, err := SpanningTreeFromObjectValue(ctx, iface.SpanningTree)
		if err != nil {
			tflog.Error(ctx, "Failed to convert SpanningTree from ObjectValue")
			return nil
		}
		cisIface.STPPortFast = st.Portfast.ValueString()
		if st.BpduGuard.ValueBool() {
			cisIface.STPBpduGuard = "enable"
		} else {
			cisIface.STPBpduGuard = "disable"
		}
	}
	return cisIface
}

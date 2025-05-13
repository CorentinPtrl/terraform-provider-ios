package models

import (
	"github.com/CorentinPtrl/cisconf"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type InterfacesDataSourceModel struct {
	Interfaces []InterfaceModel `tfsdk:"interfaces"`
}

type InterfaceModel struct {
	ID         types.String `tfsdk:"id"`
	Switchport types.String `tfsdk:"switchport"`
	//	Ips         []IpInterfaceModel `tfsdk:"ips"`
	Description types.String `tfsdk:"description"`
	Shutdown    types.Bool   `tfsdk:"shutdown"`
}

type IpInterfaceModel struct {
	Ip   types.String `tfsdk:"ip"`
	Mask types.String `tfsdk:"mask"`
}

func InterfaceFromCisconf(iface *cisconf.CiscoInterface) InterfaceModel {
	ips := make([]IpInterfaceModel, len(iface.Ips))
	for i, ip := range iface.Ips {
		ips[i] = IpInterfaceModel{
			Ip:   types.StringValue(ip.Ip),
			Mask: types.StringValue(ip.Subnet),
		}
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

	return InterfaceModel{
		ID:         types.StringValue(iface.Parent.Identifier),
		Switchport: switchport,
		//		Ips:         ips,
		Description: types.StringValue(iface.Description),
		Shutdown:    types.BoolValue(iface.Shutdown),
	}
}

func InterfaceToCisconf(iface InterfaceModel) *cisconf.CiscoInterface {
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

	/*	for _, ip := range iface.Ips {
			cisIface.Ips = append(cisIface.Ips, cisconf.Ip{
				Ip:     ip.Ip.ValueString(),
				Subnet: ip.Mask.ValueString(),
			})
		}
	*/
	return cisIface
}

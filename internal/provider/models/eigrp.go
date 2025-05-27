// Copyright (c) Corentin Pitrel
// SPDX-License-Identifier: MIT

package models

import (
	"context"
	"fmt"
	"github.com/CorentinPtrl/cisconf"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"net"
	"terraform-provider-ios/internal/utils"
)

type EigrpModel struct {
	As       types.Int64 `tfsdk:"as_number"`
	Networks types.List  `tfsdk:"networks"`
}

func EigrpToCisconf(ctx context.Context, data EigrpModel) cisconf.Eigrp {
	var networks []cisconf.EigrpNetwork
	var networkList []string
	data.Networks.ElementsAs(ctx, &networkList, true)
	for _, network := range networkList {
		_, ipNet, err := net.ParseCIDR(network)
		if err != nil {
			tflog.Info(ctx, fmt.Sprintf("Failed to parse CIDR %s: %v", network, err))
			return cisconf.Eigrp{}
		}
		subnet := ipNet.IP.String()
		mask := net.IP(ipNet.Mask).String()
		wildcard, err := utils.MaskToWildcard(mask)
		if err != nil {
			tflog.Info(ctx, fmt.Sprintf("Failed to convert mask %s to wildcard: %v", mask, err))
			return cisconf.Eigrp{}
		}
		networks = append(networks, cisconf.EigrpNetwork{
			NetworkNumber: subnet,
			WildCard:      wildcard,
		})
	}
	return cisconf.Eigrp{
		Asn:     int(data.As.ValueInt64()),
		Network: networks,
	}
}

func EigrpFromCisconf(ctx context.Context, vlan cisconf.Eigrp) EigrpModel {
	var networks []string
	for _, network := range vlan.Network {
		cidr, err := utils.WildcardToCIDR(network.WildCard)
		if err != nil {
			tflog.Info(ctx, fmt.Sprintf("Failed to convert wildcard %s to CIDR: %v", network.WildCard, err))
			return EigrpModel{}
		}
		networks = append(networks, network.NetworkNumber+"/"+fmt.Sprintf("%d", cidr))
	}
	list, diags := types.ListValueFrom(ctx, types.StringType, networks)
	if diags.HasError() {
		tflog.Error(ctx, fmt.Sprintf("Failed to convert networks to ListValue: %v", diags))
		return EigrpModel{}
	}

	return EigrpModel{
		As:       types.Int64Value(int64(vlan.Asn)),
		Networks: list,
	}
}

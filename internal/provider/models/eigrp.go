// Copyright (c) Corentin Pitrel
// SPDX-License-Identifier: MIT

package models

import (
	"context"
	"fmt"
	"github.com/CorentinPtrl/cgnet"
	"github.com/CorentinPtrl/cisconf"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"net"
	"terraform-provider-ios/internal/utils"
)

type EigrpModel struct {
	As       types.Int64 `tfsdk:"as_number"`
	Networks types.List  `tfsdk:"networks"`
}

func EigrpToCisconf(ctx context.Context, data EigrpModel) (cisconf.Eigrp, error) {
	var networks []cisconf.EigrpNetwork
	var networkList []string
	data.Networks.ElementsAs(ctx, &networkList, true)
	for _, network := range networkList {
		_, ipNet, err := net.ParseCIDR(network)
		if err != nil {
			return cisconf.Eigrp{}, fmt.Errorf("failed to parse CIDR %s: %v", network, err)
		}
		subnet := ipNet.IP.String()
		mask := net.IP(ipNet.Mask).String()
		wildcard, err := utils.MaskToWildcard(mask)
		if err != nil {
			return cisconf.Eigrp{}, fmt.Errorf("failed to convert mask %s to wildcard: %v", mask, err)
		}
		networks = append(networks, cisconf.EigrpNetwork{
			NetworkNumber: subnet,
			WildCard:      wildcard,
		})
	}
	return cisconf.Eigrp{
		Asn:     int(data.As.ValueInt64()),
		Network: networks,
	}, nil
}

func EigrpFromCisconf(ctx context.Context, vlan cisconf.Eigrp) (EigrpModel, error) {
	var networks []string
	for _, network := range vlan.Network {
		cidr, err := utils.WildcardToCIDR(network.WildCard)
		if err != nil {
			return EigrpModel{}, fmt.Errorf("failed to convert wildcard %s to CIDR: %v", network.WildCard, err)
		}
		networks = append(networks, network.NetworkNumber+"/"+fmt.Sprintf("%d", cidr))
	}
	list, diags := types.ListValueFrom(ctx, types.StringType, networks)
	if diags.HasError() {
		return EigrpModel{}, fmt.Errorf("failed to convert networks to ListValue: %v", diags)
	}

	return EigrpModel{
		As:       types.Int64Value(int64(vlan.Asn)),
		Networks: list,
	}, nil
}

func GetEigrpProcesses(ctx context.Context, device *cgnet.Device) ([]EigrpModel, error) {
	config, err := device.Exec("sh running-config")
	if err != nil {
		return nil, fmt.Errorf("failed to execute running config: %w", err)
	}
	runningConfig := cisconf.Config{}
	err = cisconf.Unmarshal(config, &runningConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal running config: %w", err)
	}
	result := []EigrpModel{}
	for _, eigrp := range runningConfig.EIGRPProcess {
		var eigrpModel EigrpModel
		eigrpModel, err = EigrpFromCisconf(ctx, eigrp)
		if err != nil {
			return nil, fmt.Errorf("failed to convert EIGRP: %w", err)
		}
		result = append(result, eigrpModel)
	}
	return result, nil
}

func GetEigrpProcess(ctx context.Context, device *cgnet.Device, asn int64) (*EigrpModel, error) {
	eigrpProcesses, err := GetEigrpProcesses(ctx, device)
	if err != nil {
		return nil, fmt.Errorf("failed to get EIGRP processes: %w", err)
	}
	for _, eigrp := range eigrpProcesses {
		if eigrp.As.ValueInt64() == asn {
			return &eigrp, nil
		}
	}
	return nil, fmt.Errorf("EIGRP process with ASN %d not found", asn)
}

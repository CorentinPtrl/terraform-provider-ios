// Copyright (c) Corentin Pitrel
// SPDX-License-Identifier: MIT

package models

import (
	"context"
	"fmt"
	"github.com/CorentinPtrl/cgnet"
	"github.com/CorentinPtrl/cisconf"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sirikothe/gotextfsm"
	"strconv"
	"terraform-provider-ios/internal/provider/ntc"
)

type VlanModel struct {
	Id   types.Int32  `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

type VlansDataSourceModel struct {
	Vlans []VlanModel `tfsdk:"vlans"`
}

func VlanToCisconf(ctx context.Context, data VlanModel) cisconf.Vlan {
	return cisconf.Vlan{
		Id:   int(data.Id.ValueInt32()),
		Name: data.Name.ValueString(),
	}
}

func VlanFromCisconf(ctx context.Context, vlan cisconf.Vlan) VlanModel {
	return VlanModel{
		Id:   types.Int32Value(int32(vlan.Id)),
		Name: types.StringValue(vlan.Name),
	}
}
func GetVlans(device *cgnet.Device) (map[int]VlanModel, error) {
	config, err := device.Exec("sh running-config")
	if err != nil {
		return nil, fmt.Errorf("failed to execute running config: %w", err)
	}
	runningConfig := cisconf.Config{}
	err = cisconf.Unmarshal(config, &runningConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal running config: %w", err)
	}

	vlans := map[int]VlanModel{}
	for _, v := range runningConfig.Vlans {
		vlans[v.Id] = VlanModel{
			Id:   types.Int32Value(int32(v.Id)),
			Name: types.StringValue(v.Name),
		}
	}

	fsm, err := ntc.GetTextFSM("cisco_ios_show_vlan.textfsm")
	if err != nil {
		return nil, fmt.Errorf("failed to get textfsm: %w", err)
	}
	result, err := device.Exec("sh vlan")
	if err != nil {
		return nil, fmt.Errorf("failed to execute running config: %w", err)
	}
	parser := gotextfsm.ParserOutput{}
	err = parser.ParseTextString(result, fsm, false)
	if err != nil {
		return nil, fmt.Errorf("failed to parse result: %w", err)
	}
	for _, dic := range parser.Dict {
		id, err := strconv.Atoi(dic["VLAN_ID"].(string))
		if err != nil {
			return nil, fmt.Errorf("failed to parse VLAN ID: %w", err)
		}
		vlans[id] = VlanModel{
			Id:   types.Int32Value(int32(id)),
			Name: types.StringValue(dic["VLAN_NAME"].(string)),
		}
	}
	return vlans, nil
}

func GetVlan(device *cgnet.Device, id int) (*VlanModel, error) {
	vlans, err := GetVlans(device)
	if err != nil {
		return nil, fmt.Errorf("failed to get VLANs: %w", err)
	}
	vlan, ok := vlans[id]
	if !ok {
		return nil, nil
	}
	return &vlan, nil
}

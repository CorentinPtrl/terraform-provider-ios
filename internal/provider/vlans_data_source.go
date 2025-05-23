// Copyright (c) Corentin Pitrel
// SPDX-License-Identifier: MIT

package provider

import (
	"context"
	"fmt"
	"github.com/CorentinPtrl/cisconf"
	"github.com/Letsu/cgnet"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sirikothe/gotextfsm"
	"sort"
	"strconv"
	"terraform-provider-ios/internal/provider/models"
	"terraform-provider-ios/internal/provider/ntc"
)

var _ datasource.DataSource = &VlansDataSource{}

func NewVlansDataSource() datasource.DataSource {
	return &VlansDataSource{}
}

type VlansDataSource struct {
	client *cgnet.Device
}

func (d *VlansDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vlans"
}

func (d *VlansDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Vlans data source",

		Attributes: map[string]schema.Attribute{
			"vlans": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.Int64Attribute{
							Computed: true,
						},
						"name": schema.StringAttribute{
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func (d *VlansDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*cgnet.Device)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.client = client
}

func (d *VlansDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data models.VlansDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	config, err := d.client.Exec("sh running-config")
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to execute running config",
			fmt.Sprintf("Unable to execute running config: %s", err),
		)
		return
	}
	runningConfig := cisconf.Config{}
	err = cisconf.Unmarshal(config, &runningConfig)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to unmarshal running config",
			fmt.Sprintf("Unable to parse running config: %s", err),
		)
		return
	}

	vlans := map[int]models.VlanModel{}
	for _, v := range runningConfig.Vlans {
		vlans[v.Id] = models.VlanModel{
			Id:   types.Int32Value(int32(v.Id)),
			Name: types.StringValue(v.Name),
		}
	}

	fsm, err := ntc.GetTextFSM("cisco_ios_show_vlan.textfsm")
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to get textfsm",
			fmt.Sprintf("Unable to get textfsm: %s", err),
		)
		return
	}
	result, err := d.client.Exec("sh vlan")
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to execute running config",
			fmt.Sprintf("Unable to execute running config: %s", err),
		)
		return
	}
	parser := gotextfsm.ParserOutput{}
	err = parser.ParseTextString(result, fsm, false)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to parse result",
			fmt.Sprintf("Unable to parse result: %s", err),
		)
		return
	}
	for _, dic := range parser.Dict {
		id, err := strconv.Atoi(dic["VLAN_ID"].(string))
		if err != nil {
			resp.Diagnostics.AddError(
				"Failed to parse VLAN ID",
				fmt.Sprintf("Unable to parse VLAN ID: %s", err),
			)
			return
		}
		vlans[id] = models.VlanModel{
			Id:   types.Int32Value(int32(id)),
			Name: types.StringValue(dic["VLAN_NAME"].(string)),
		}
	}

	keys := make([]int, 0, len(vlans))
	for k := range vlans {
		keys = append(keys, k)
	}
	sort.Ints(keys)

	for _, k := range keys {
		data.Vlans = append(data.Vlans, vlans[k])
	}
	
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

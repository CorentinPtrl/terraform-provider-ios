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
	"github.com/sirikothe/gotextfsm"
	"terraform-provider-ios/internal/provider/models"
	"terraform-provider-ios/internal/provider/ntc"
)

var _ datasource.DataSource = &VlanDataSource{}

func NewVlanDataSource() datasource.DataSource {
	return &VlanDataSource{}
}

type VlanDataSource struct {
	client *cgnet.Device
}

func (d *VlanDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vlan"
}

func (d *VlanDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Vlan data source",

		Attributes: map[string]schema.Attribute{
			"id": schema.Int32Attribute{
				Required: true,
			},
			"name": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (d *VlanDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *VlanDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data models.VlanModel

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

	var vlan *cisconf.Vlan
	vlan = nil
	for _, v := range runningConfig.Vlans {
		if v.Id == int(data.Id.ValueInt32()) {
			vlan = &v
			break
		}
	}

	if vlan == nil {
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
			if dic["VLAN_ID"] == fmt.Sprintf("%d", data.Id.ValueInt32()) {
				vlan = &cisconf.Vlan{
					Id:   int(data.Id.ValueInt32()),
					Name: dic["VLAN_NAME"].(string),
				}
				break
			}
		}
	}

	if vlan == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	data = models.VlanFromCisconf(ctx, *vlan)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

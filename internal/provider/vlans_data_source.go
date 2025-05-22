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
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ datasource.DataSource = &VlansDataSource{}

func NewVlansDataSource() datasource.DataSource {
	return &VlansDataSource{}
}

type VlansDataSource struct {
	client *cgnet.Device
}

type VlansDataSourceModel struct {
	Vlans []vlanModel `tfsdk:"vlans"`
}

type vlanModel struct {
	ID   types.Int64  `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
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
	var data VlansDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	config, err := d.client.Exec("do show running-config")
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to read command",
			fmt.Sprintf("Unable to execute command: %s", err),
		)
		return
	}
	var runningConfig cisconf.Config
	err = cisconf.Unmarshal(config, &runningConfig)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to unmarshal running config",
			fmt.Sprintf("Unable to parse running config: %s", err),
		)
		return
	}
	for _, vlan := range runningConfig.Vlans {
		vlanModel := vlanModel{
			ID:   types.Int64Value(int64(vlan.Id)),
			Name: types.StringValue(vlan.Name),
		}
		data.Vlans = append(data.Vlans, vlanModel)
	}

	tflog.Info(ctx, fmt.Sprintf("Vlans: %v", runningConfig.Vlans))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

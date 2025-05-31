// Copyright (c) Corentin Pitrel
// SPDX-License-Identifier: MIT

package provider

import (
	"context"
	"fmt"
	"github.com/CorentinPtrl/cgnet"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"terraform-provider-ios/internal/provider/models"
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

	vlans, err := models.GetVlans(d.client)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to get VLANs",
			fmt.Sprintf("An error occurred while retrieving VLANs: %s", err),
		)
		return
	}

	vlan, ok := vlans[int(data.Id.ValueInt32())]
	if !ok {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &vlan)...)
}

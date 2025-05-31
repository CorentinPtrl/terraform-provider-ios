// Copyright (c) Corentin Pitrel
// SPDX-License-Identifier: MIT

package provider

import (
	"context"
	"fmt"
	"github.com/CorentinPtrl/cgnet"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"sort"
	"terraform-provider-ios/internal/provider/models"
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

	vlans, err := models.GetVlans(d.client)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to get VLANs",
			fmt.Sprintf("Unable to retrieve VLANs: %s", err),
		)
		return
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

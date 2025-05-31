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

var _ datasource.DataSource = &StaticRoutesDataSource{}

func NewStaticRoutesDataSource() datasource.DataSource {
	return &StaticRoutesDataSource{}
}

type StaticRoutesDataSource struct {
	client *cgnet.Device
}

func (d *StaticRoutesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_static_routes"
}

func (d *StaticRoutesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Static Routes data source",

		Attributes: map[string]schema.Attribute{
			"routes": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"prefix": schema.StringAttribute{
							Computed: true,
						},
						"mask": schema.StringAttribute{
							Computed: true,
						},
						"next_hop": schema.StringAttribute{
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func (d *StaticRoutesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *StaticRoutesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data models.RoutesDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	routes, err := models.GetRoutes(d.client)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to get static routes",
			fmt.Sprintf("An error occurred while retrieving static routes: %s", err),
		)
		return
	}
	data.Routes = routes
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

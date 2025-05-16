// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"github.com/CorentinPtrl/cisconf"
	"github.com/Letsu/cgnet"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"terraform-provider-ios/internal/provider/models"
)

var _ datasource.DataSource = &InterfacesDataSource{}

func NewInterfacesDataSource() datasource.DataSource {
	return &InterfacesDataSource{}
}

type InterfacesDataSource struct {
	client *cgnet.Device
}

func (d *InterfacesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_interfaces"
}

func (d *InterfacesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Interfaces data source",

		Attributes: map[string]schema.Attribute{
			"interfaces": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed: true,
						},
						"switchport": schema.StringAttribute{
							Computed: true,
						},
						"spanning_tree": schema.ObjectAttribute{
							AttributeTypes: map[string]attr.Type{
								"portfast":   types.StringType,
								"bpdu_guard": types.BoolType,
							},
							Computed: true,
							Optional: true,
						},
						"trunk": schema.ObjectAttribute{
							AttributeTypes: map[string]attr.Type{
								"encapsulation": types.StringType,
								"allowed_vlans": types.ListType{}.WithElementType(types.Int32Type),
							},
							Computed: true,
							Optional: true,
						},
						"access": schema.ObjectAttribute{
							AttributeTypes: map[string]attr.Type{
								"access_vlan": types.Int32Type,
							},
							Computed: true,
							Optional: true,
						},
						"description": schema.StringAttribute{
							Computed: true,
						},
						"shutdown": schema.BoolAttribute{
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func (d *InterfacesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *InterfacesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data models.InterfacesSwitchesDataSourceModel

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
	var runningConfig cisconf.Config
	err = cisconf.Unmarshal(config, &runningConfig)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to unmarshal running config",
			fmt.Sprintf("Unable to parse running config: %s", err),
		)
		return
	}
	for _, inter := range runningConfig.Interfaces {
		data.Interfaces = append(data.Interfaces, models.InterfaceSwitchFromCisconf(ctx, &inter))
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

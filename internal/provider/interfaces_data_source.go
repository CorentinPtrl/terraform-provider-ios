// Copyright (c) Corentin Pitrel
// SPDX-License-Identifier: MIT

package provider

import (
	"context"
	"fmt"
	"github.com/CorentinPtrl/cgnet"
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
							Computed:    true,
							Description: "The unique identifier for the interface, typically in the format 'GigabitEthernet0/1'.",
						},
						"switchport": schema.StringAttribute{
							Computed:    true,
							Description: "The switchport mode of the interface, such as 'access' or 'trunk'. If not set, the interface is assumed to be in routed mode.",
						},
						"spanning_tree": schema.ObjectAttribute{
							AttributeTypes: map[string]attr.Type{
								"portfast":   types.StringType,
								"bpdu_guard": types.BoolType,
							},
							Computed:    true,
							Optional:    true,
							Description: "Spanning Tree Protocol (STP) settings for the interface. 'portfast' enables PortFast, and 'bpdu_guard' enables BPDU Guard.",
						},
						"trunk": schema.ObjectAttribute{
							AttributeTypes: map[string]attr.Type{
								"encapsulation": types.StringType,
								"allowed_vlans": types.ListType{}.WithElementType(types.Int32Type),
							},
							Computed:    true,
							Optional:    true,
							Description: "Trunk settings for the interface. 'encapsulation' specifies the trunk encapsulation type, and 'allowed_vlans' lists the VLANs allowed on the trunk.",
						},
						"access": schema.ObjectAttribute{
							AttributeTypes: map[string]attr.Type{
								"access_vlan": types.Int32Type,
							},
							Computed:    true,
							Optional:    true,
							Description: "Access settings for the interface. 'access_vlan' specifies the VLAN assigned to the access port.",
						},
						"description": schema.StringAttribute{
							Computed: true,
						},
						"shutdown": schema.BoolAttribute{
							Computed:    true,
							Description: "Indicates whether the interface is administratively shut down. If true, the interface is disabled.",
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

	interfaces, err := models.GetSwitchInterfaces(ctx, d.client)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to get interfaces",
			fmt.Sprintf("An error occurred while retrieving interfaces: %s", err),
		)
		return
	}
	data.Interfaces = interfaces

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

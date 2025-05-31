// Copyright (c) Corentin Pitrel
// SPDX-License-Identifier: MIT

package provider

import (
	"context"
	"fmt"
	"github.com/CorentinPtrl/cgnet"
	"github.com/CorentinPtrl/cisconf"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"terraform-provider-ios/internal/provider/models"
	"terraform-provider-ios/internal/utils"
)

var _ resource.Resource = &InterfaceEthernetResource{}

func NewInterfaceEthernetResource() resource.Resource {
	return &InterfaceEthernetResource{}
}

type InterfaceEthernetResource struct {
	client *cgnet.Device
}

func (r *InterfaceEthernetResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ethernet_interface"
}

func (r *InterfaceEthernetResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	defaultIpList, diags := types.ListValueFrom(ctx, types.ObjectType{}.WithAttributeTypes(models.IpInterfaceModel{}.AttributeTypes()), []models.IpInterfaceModel{})
	if diags.HasError() {
		resp.Diagnostics.AddError(
			"Failed to create default IP defaultIpList",
			fmt.Sprintf("Unable to create default IP defaultIpList: %s", diags),
		)
		return
	}
	resp.Schema = schema.Schema{
		MarkdownDescription: "Switch Interface resource",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the interface, e.g., 'GigabitEthernet0/1'.",
			},
			"ips": schema.ListNestedAttribute{
				Computed: true,
				Optional: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"ip": schema.StringAttribute{
							Required: true,
						},
					},
				},
				Default:     listdefault.StaticValue(defaultIpList),
				Description: "List of IP addresses assigned to the interface. Each IP address must be specified in CIDR notation (e.g., '192.168.10.2/24').",
			},
			"helper_addresses": schema.ListAttribute{
				Computed:    true,
				Optional:    true,
				ElementType: types.StringType,
				Default:     listdefault.StaticValue(types.ListNull(types.StringType)),
				Description: "List of helper addresses for the interface. These addresses are used for protocols like DHCP and TFTP to forward requests to the appropriate server.",
			},
			"description": schema.StringAttribute{
				Computed:    true,
				Optional:    true,
				Default:     stringdefault.StaticString(""),
				Description: "Description of the interface.",
			},
			"shutdown": schema.BoolAttribute{
				Computed:    true,
				Optional:    true,
				Default:     booldefault.StaticBool(false),
				Description: "Indicates whether the interface is administratively shut down. If true, the interface is disabled.",
			},
		},
	}
}

func (r *InterfaceEthernetResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*cgnet.Device)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

func (r *InterfaceEthernetResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data models.InterfaceEthernetModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	inter, err := models.GetEthernetInterface(ctx, r.client, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to get ethernet interface",
			fmt.Sprintf("Unable to get ethernet interface: %s", err),
		)
		return
	}

	var ethernetConfig *cisconf.CiscoInterface
	ethernetConfig, err = models.InterfaceEthernetToCisconf(ctx, data)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to convert interface model to CISCONF",
			fmt.Sprintf("Unable to convert interface model: %s", err),
		)
		return
	}
	interCisco, err := models.InterfaceEthernetToCisconf(ctx, inter)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to convert interface model to CISCONF",
			fmt.Sprintf("Unable to convert interface model: %s", err),
		)
		return
	}
	marshal, err := cisconf.Diff(*interCisco, *ethernetConfig)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to diff interface",
			fmt.Sprintf("Unable to diff interface: %s", err),
		)
		return
	}
	err = utils.ConfigDevice(marshal, r.client)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to configure interface",
			fmt.Sprintf("Unable to configure interface: %s", err),
		)
		return
	}

	inter, err = models.GetEthernetInterface(ctx, r.client, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to get ethernet interface",
			fmt.Sprintf("Unable to get ethernet interface: %s", err),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &inter)...)
}

func (r *InterfaceEthernetResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data models.InterfaceEthernetModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	inter, err := models.GetEthernetInterface(ctx, r.client, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to get ethernet interface",
			fmt.Sprintf("Unable to get ethernet interface: %s", err),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &inter)...)
}

func (r *InterfaceEthernetResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data models.InterfaceEthernetModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	inter, err := models.GetEthernetInterface(ctx, r.client, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to get ethernet interface",
			fmt.Sprintf("Unable to get ethernet interface: %s", err),
		)
		return
	}

	var ethernetConfig *cisconf.CiscoInterface
	ethernetConfig, err = models.InterfaceEthernetToCisconf(ctx, data)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to convert interface model to CISCONF",
			fmt.Sprintf("Unable to convert interface model: %s", err),
		)
		return
	}
	interCisco, err := models.InterfaceEthernetToCisconf(ctx, inter)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to convert interface model to CISCONF",
			fmt.Sprintf("Unable to convert interface model: %s", err),
		)
		return
	}
	marshal, err := cisconf.Diff(*interCisco, *ethernetConfig)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to diff interface",
			fmt.Sprintf("Unable to diff interface: %s", err),
		)
		return
	}
	err = utils.ConfigDevice(marshal, r.client)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to configure interface",
			fmt.Sprintf("Unable to configure interface: %s", err),
		)
		return
	}

	inter, err = models.GetEthernetInterface(ctx, r.client, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to get ethernet interface",
			fmt.Sprintf("Unable to get ethernet interface: %s", err),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &inter)...)
}

func (r *InterfaceEthernetResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data models.InterfaceEthernetModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Configure([]string{"default interface " + data.ID.ValueString()})
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to configure interface",
			fmt.Sprintf("Unable to configure interface: %s", err),
		)
		return
	}
}

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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"terraform-provider-ios/internal/provider/models"
	"terraform-provider-ios/internal/utils"
)

var _ resource.Resource = &InterfaceSwitchResource{}

func NewInterfaceSwitchResource() resource.Resource {
	return &InterfaceSwitchResource{}
}

type InterfaceSwitchResource struct {
	client *cgnet.Device
}

func (r *InterfaceSwitchResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_switch_interface"
}

func (r *InterfaceSwitchResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	st_obj, diags := types.ObjectValue(models.DefaultSpanningTree.AttributeTypes(), models.DefaultSpanningTree.AttributeValues())
	if diags.HasError() {
		resp.Diagnostics.AddError(
			"Failed to create default spanning tree object",
			fmt.Sprintf("Unable to create default spanning tree object: %s", diags),
		)
		return
	}
	resp.Schema = schema.Schema{
		MarkdownDescription: "Switch Interface resource",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required: true,
			},
			"switchport": schema.StringAttribute{
				Computed: true,
			},
			"trunk": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"encapsulation": schema.StringAttribute{
						MarkdownDescription: "Encapsulation type",
						Optional:            true,
						Computed:            true,
						Default:             stringdefault.StaticString("dot1q"),
					},
					"allowed_vlans": schema.ListAttribute{
						MarkdownDescription: "Allowed VLANs",
						ElementType:         types.Int32Type,
						Optional:            true,
						Computed:            true,
						Default:             listdefault.StaticValue(types.ListNull(types.Int32Type)),
					},
				},
				MarkdownDescription: "Trunk configuration",
				Optional:            true,
				Computed:            true,
				Default:             objectdefault.StaticValue(types.ObjectNull(models.Trunk{}.AttributeTypes())),
			},
			"access": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"access_vlan": schema.Int32Attribute{
						MarkdownDescription: "Access VLAN",
						Optional:            true,
						Computed:            true,
						Default:             int32default.StaticInt32(1),
					},
				},
				Optional: true,
				Computed: true,
				Default:  objectdefault.StaticValue(types.ObjectNull(models.Access{}.AttributeTypes())),
			},
			"spanning_tree": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"portfast": schema.StringAttribute{
						Optional: true,
						Computed: true,
						Default:  stringdefault.StaticString(""),
					},
					"bpdu_guard": schema.BoolAttribute{
						Optional: true,
					},
				},
				Optional: true,
				Computed: true,
				Default:  objectdefault.StaticValue(st_obj),
			},
			"description": schema.StringAttribute{
				Computed: true,
				Optional: true,
				Default:  stringdefault.StaticString(""),
			},
			"shutdown": schema.BoolAttribute{
				Computed: true,
				Optional: true,
				Default:  booldefault.StaticBool(false),
			},
		},
	}
}

func (r *InterfaceSwitchResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *InterfaceSwitchResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data models.InterfaceSwitchModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	inter, err := models.GetSwitchInterface(ctx, r.client, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to get switch interface",
			fmt.Sprintf("Unable to get switch interface: %s", err),
		)
		return
	}

	var interfaceSwitch *cisconf.CiscoInterface
	interfaceSwitch, err = models.InterfaceSwitchToCisconf(ctx, data)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to convert interface switch model",
			fmt.Sprintf("Unable to convert interface switch model: %s", err),
		)
		return
	}
	interCisco, err := models.InterfaceSwitchToCisconf(ctx, inter)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to convert interface switch model",
			fmt.Sprintf("Unable to convert interface switch model: %s", err),
		)
		return
	}
	marshal, err := cisconf.Diff(*interCisco, *interfaceSwitch)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to diff interface switch",
			fmt.Sprintf("Unable to diff interface switch: %s", err),
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

	inter, err = models.GetSwitchInterface(ctx, r.client, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to get switch interface",
			fmt.Sprintf("Unable to get switch interface: %s", err),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &inter)...)
}

func (r *InterfaceSwitchResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data models.InterfaceSwitchModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	inter, err := models.GetSwitchInterface(ctx, r.client, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to get switch interface",
			fmt.Sprintf("Unable to get switch interface: %s", err),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &inter)...)
}

func (r *InterfaceSwitchResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data models.InterfaceSwitchModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	inter, err := models.GetSwitchInterface(ctx, r.client, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to get switch interface",
			fmt.Sprintf("Unable to get switch interface: %s", err),
		)
		return
	}
	var interfaceSwitch *cisconf.CiscoInterface
	interfaceSwitch, err = models.InterfaceSwitchToCisconf(ctx, data)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to convert interface switch model",
			fmt.Sprintf("Unable to convert interface switch model: %s", err),
		)
		return
	}
	interCisco, err := models.InterfaceSwitchToCisconf(ctx, inter)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to convert interface switch model",
			fmt.Sprintf("Unable to convert interface switch model: %s", err),
		)
		return
	}
	marshal, err := cisconf.Diff(*interCisco, *interfaceSwitch)
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

	inter, err = models.GetSwitchInterface(ctx, r.client, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to get switch interface",
			fmt.Sprintf("Unable to get switch interface: %s", err),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &inter)...)
}

func (r *InterfaceSwitchResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data models.InterfaceSwitchModel

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

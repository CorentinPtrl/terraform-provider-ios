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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"terraform-provider-ios/internal/provider/models"
	"terraform-provider-ios/internal/utils"
)

var _ resource.Resource = &VlanResource{}

func NewVlanResource() resource.Resource {
	return &VlanResource{}
}

type VlanResource struct {
	client *cgnet.Device
}

func (r *VlanResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vlan"
}

func (r *VlanResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Vlan resource",

		Attributes: map[string]schema.Attribute{
			"id": schema.Int32Attribute{
				Required: true,
				PlanModifiers: []planmodifier.Int32{
					int32planmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Required: true,
			},
		},
	}
}

func (r *VlanResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *VlanResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data models.VlanModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	vlan, err := models.GetVlan(r.client, int(data.Id.ValueInt32()))
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to get vlan",
			fmt.Sprintf("Unable to get vlan: %s", err),
		)
		return
	}

	var marshal string
	if vlan == nil {
		marshal, err = cisconf.Marshal(models.VlanToCisconf(ctx, data))
	} else {
		marshal, err = cisconf.Diff(models.VlanToCisconf(ctx, *vlan), models.VlanToCisconf(ctx, data))
	}
	if err != nil {
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

	vlan, err = models.GetVlan(r.client, int(data.Id.ValueInt32()))
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to get vlan",
			fmt.Sprintf("Unable to get vlan: %s", err),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, vlan)...)
}

func (r *VlanResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data models.VlanModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	vlan, err := models.GetVlan(r.client, int(data.Id.ValueInt32()))
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to get vlan",
			fmt.Sprintf("Unable to get vlan: %s", err),
		)
		return
	}

	if vlan == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, vlan)...)
}

func (r *VlanResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data models.VlanModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	vlan, err := models.GetVlan(r.client, int(data.Id.ValueInt32()))
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to get vlan",
			fmt.Sprintf("Unable to get vlan: %s", err),
		)
		return
	}

	var marshal string
	if vlan == nil {
		marshal, err = cisconf.Marshal(models.VlanToCisconf(ctx, data))
	} else {
		marshal, err = cisconf.Diff(models.VlanToCisconf(ctx, *vlan), models.VlanToCisconf(ctx, data))
	}
	if err != nil {
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

	vlan, err = models.GetVlan(r.client, int(data.Id.ValueInt32()))
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to get vlan",
			fmt.Sprintf("Unable to get vlan: %s", err),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, vlan)...)
}

func (r *VlanResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data models.VlanModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Configure([]string{"no vlan " + fmt.Sprintf("%d", data.Id.ValueInt32())})
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to configure interface",
			fmt.Sprintf("Unable to configure interface: %s", err),
		)
		return
	}
}

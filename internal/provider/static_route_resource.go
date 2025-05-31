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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"terraform-provider-ios/internal/provider/models"
	"terraform-provider-ios/internal/utils"
)

var _ resource.Resource = &StaticRouteResource{}

func NewStaticRouteResource() resource.Resource {
	return &StaticRouteResource{}
}

type StaticRouteResource struct {
	client *cgnet.Device
}

func (r *StaticRouteResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_static_route"
}

func (r *StaticRouteResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Static Route resource",

		Attributes: map[string]schema.Attribute{
			"prefix": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"mask": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"next_hop": schema.StringAttribute{
				Required: true,
			},
		},
	}
}

func (r *StaticRouteResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *StaticRouteResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data models.RouteModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	routes, err := models.GetRoutes(r.client)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to get static routes",
			fmt.Sprintf("An error occurred while retrieving static routes: %s", err),
		)
		return
	}

	var route *models.RouteModel
	route = nil
	for _, v := range routes {
		if v.Prefix == data.Prefix && v.Mask == data.Mask {
			route = &v
			break
		}
	}

	var marshal string
	if route == nil {
		dest := cisconf.RoutesType{Routes: []cisconf.Route{
			models.RouteToCisconf(data),
		}}
		marshal, err = cisconf.Marshal(dest)
	} else {
		src := cisconf.RoutesType{Routes: []cisconf.Route{
			models.RouteToCisconf(*route),
		}}
		dest := cisconf.RoutesType{Routes: []cisconf.Route{
			models.RouteToCisconf(data),
		}}
		marshal, err = cisconf.Diff(src, dest)
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

	routes, err = models.GetRoutes(r.client)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to get static routes",
			fmt.Sprintf("An error occurred while retrieving static routes: %s", err),
		)
		return
	}
	route = nil
	for _, v := range routes {
		if v.Prefix == data.Prefix && v.Mask == data.Mask {
			route = &v
			break
		}
	}

	if route == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, route)...)
}

func (r *StaticRouteResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data models.RouteModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	routes, err := models.GetRoutes(r.client)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to get static routes",
			fmt.Sprintf("An error occurred while retrieving static routes: %s", err),
		)
		return
	}
	var route *models.RouteModel
	route = nil
	for _, v := range routes {
		if v.Prefix == data.Prefix && v.Mask == data.Mask {
			route = &v
			break
		}
	}

	if route == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &route)...)
}

func (r *StaticRouteResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data models.RouteModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	routes, err := models.GetRoutes(r.client)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to get static routes",
			fmt.Sprintf("An error occurred while retrieving static routes: %s", err),
		)
		return
	}
	var route *models.RouteModel
	route = nil
	for _, v := range routes {
		if v.Prefix == data.Prefix && v.Mask == data.Mask {
			route = &v
			break
		}
	}

	var marshal string
	if route == nil {
		dest := cisconf.RoutesType{Routes: []cisconf.Route{
			models.RouteToCisconf(data),
		}}
		marshal, err = cisconf.Marshal(dest)
	} else {
		src := cisconf.RoutesType{Routes: []cisconf.Route{
			models.RouteToCisconf(*route),
		}}
		dest := cisconf.RoutesType{Routes: []cisconf.Route{
			models.RouteToCisconf(data),
		}}
		marshal, err = cisconf.Diff(src, dest)
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
	routes, err = models.GetRoutes(r.client)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to get static routes",
			fmt.Sprintf("An error occurred while retrieving static routes: %s", err),
		)
		return
	}
	route = nil
	for _, v := range routes {
		if v.Prefix == data.Prefix && v.Mask == data.Mask {
			route = &v
			break
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, route)...)
}

func (r *StaticRouteResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data models.RouteModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Configure([]string{"no ip route " + data.Prefix.ValueString() + " " + data.Mask.ValueString() + " " + data.NextHop.ValueString()})
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to configure interface",
			fmt.Sprintf("Unable to configure interface: %s", err),
		)
		return
	}
}

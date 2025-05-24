// Copyright (c) Corentin Pitrel
// SPDX-License-Identifier: MIT

package provider

import (
	"context"
	"fmt"
	"github.com/CorentinPtrl/cisconf"
	"github.com/Letsu/cgnet"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"strings"
	"terraform-provider-ios/internal/provider/models"
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

	config, err := r.client.Exec("sh running-config")
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

	var route *cisconf.Route
	route = nil
	for _, v := range runningConfig.Routes {
		if v.Prefix == data.Prefix.ValueString() && v.Mask == data.Mask.ValueString() {
			route = &v
			break
		}
	}

	var marshal string
	if route == nil {
		dest := cisconf.Routes{Routes: []cisconf.Route{
			models.RouteToCisconf(ctx, data),
		}}
		marshal, err = cisconf.Marshal(dest)
	} else {
		src := cisconf.Routes{Routes: []cisconf.Route{
			*route,
		}}
		dest := cisconf.Routes{Routes: []cisconf.Route{
			models.RouteToCisconf(ctx, data),
		}}
		marshal, err = cisconf.Diff(src, dest)
	}
	if err != nil {
		return
	}
	lines := strings.Split(string(marshal), "\n")
	configs := []string{}
	for _, line := range lines {
		cmd := strings.Trim(line, " ")
		if strings.Contains(cmd, "!") {
			continue
		}
		configs = append(configs, cmd)
	}
	tflog.Info(ctx, marshal)
	err = r.client.Configure(configs)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to configure interface",
			fmt.Sprintf("Unable to configure interface: %s", err),
		)
		return
	}

	config, err = r.client.Exec("sh running-config")
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to execute running config",
			fmt.Sprintf("Unable to execute running config: %s", err),
		)
		return
	}
	runningConfig = cisconf.Config{}
	err = cisconf.Unmarshal(config, &runningConfig)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to unmarshal running config",
			fmt.Sprintf("Unable to parse running config: %s", err),
		)
		return
	}

	route = nil
	for _, v := range runningConfig.Routes {
		if v.Prefix == data.Prefix.ValueString() && v.Mask == data.Mask.ValueString() {
			route = &v
			break
		}
	}

	data = models.RouteFromCisconf(ctx, *route)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *StaticRouteResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data models.RouteModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	config, err := r.client.Exec("sh running-config")
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to execute running config",
			fmt.Sprintf("Unable to execute running config: %s", err),
		)
		return
	}
	runningConfig := cisconf.Config{}
	err = cisconf.Unmarshal(config, &runningConfig)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to unmarshal running config",
			fmt.Sprintf("Unable to parse running config: %s", err),
		)
		return
	}

	var route *cisconf.Route
	route = nil
	for _, v := range runningConfig.Routes {
		if v.Prefix == data.Prefix.ValueString() && v.Mask == data.Mask.ValueString() {
			route = &v
			break
		}
	}

	if route == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	data = models.RouteFromCisconf(ctx, *route)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *StaticRouteResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data models.RouteModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}
	config, err := r.client.Exec("sh running-config")
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

	var route *cisconf.Route
	route = nil
	for _, v := range runningConfig.Routes {
		if v.Prefix == data.Prefix.ValueString() && v.Mask == data.Mask.ValueString() {
			route = &v
			break
		}
	}

	var marshal string
	if route == nil {
		dest := cisconf.Routes{Routes: []cisconf.Route{
			models.RouteToCisconf(ctx, data),
		}}
		marshal, err = cisconf.Marshal(dest)
	} else {
		src := cisconf.Routes{Routes: []cisconf.Route{
			*route,
		}}
		dest := cisconf.Routes{Routes: []cisconf.Route{
			models.RouteToCisconf(ctx, data),
		}}
		marshal, err = cisconf.Diff(src, dest)
	}
	if err != nil {
		return
	}
	lines := strings.Split(string(marshal), "\n")
	configs := []string{}
	for _, line := range lines {
		cmd := strings.Trim(line, " ")
		if strings.Contains(cmd, "!") {
			continue
		}
		configs = append(configs, cmd)
	}
	tflog.Info(ctx, marshal)
	err = r.client.Configure(configs)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to configure interface",
			fmt.Sprintf("Unable to configure interface: %s", err),
		)
		return
	}

	config, err = r.client.Exec("sh running-config")
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to execute running config",
			fmt.Sprintf("Unable to execute running config: %s", err),
		)
		return
	}
	runningConfig = cisconf.Config{}
	err = cisconf.Unmarshal(config, &runningConfig)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to unmarshal running config",
			fmt.Sprintf("Unable to parse running config: %s", err),
		)
		return
	}

	route = nil
	for _, v := range runningConfig.Routes {
		if v.Prefix == data.Prefix.ValueString() && v.Mask == data.Mask.ValueString() {
			route = &v
			break
		}
	}

	data = models.RouteFromCisconf(ctx, *route)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
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

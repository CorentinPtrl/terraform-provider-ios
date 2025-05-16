// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"github.com/CorentinPtrl/cisconf"
	"github.com/Letsu/cgnet"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"strings"
	"terraform-provider-ios/internal/provider/models"
)

var _ resource.Resource = &InterfaceResource{}

func NewInterfaceResource() resource.Resource {
	return &InterfaceResource{}
}

type InterfaceResource struct {
	client *cgnet.Device
}

func (r *InterfaceResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_switch_interface"
}

func (r *InterfaceResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Switch Interface resource",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required: true,
			},
			"switchport": schema.StringAttribute{
				Computed: true,
				Optional: true,
			},
			"encapsulation": schema.StringAttribute{
				Computed: true,
				Optional: true,
			},
			"access_vlan": schema.Int32Attribute{
				Optional: true,
				Computed: true,
			},
			"allowed_vlans": schema.ListAttribute{
				Computed:    true,
				Optional:    true,
				ElementType: types.Int32Type,
			},
			"spanning_tree_portfast": schema.StringAttribute{
				Computed: true,
				Optional: true,
			},
			"spanning_tree_bpdu_guard": schema.BoolAttribute{
				Computed: true,
				Optional: true,
			},
			"description": schema.StringAttribute{
				Computed: true,
				Optional: true,
			},
			"shutdown": schema.BoolAttribute{
				Computed: true,
				Optional: true,
			},
		},
	}
}

func (r *InterfaceResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *InterfaceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data models.InterfaceSwitchModel

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

	var inter cisconf.CiscoInterface
	for _, iface := range runningConfig.Interfaces {
		if iface.Parent.Identifier == data.ID.ValueString() {
			inter = iface
			break
		}
	}

	marshal, err := cisconf.Diff(inter, *models.InterfaceSwitchToCisconf(ctx, data))
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

	for _, iface := range runningConfig.Interfaces {
		if iface.Parent.Identifier == data.ID.ValueString() {
			inter = iface
			break
		}
	}

	data = models.InterfaceSwitchFromCisconf(ctx, &inter)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *InterfaceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data models.InterfaceSwitchModel

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
	var runningConfig cisconf.Config
	err = cisconf.Unmarshal(config, &runningConfig)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to unmarshal running config",
			fmt.Sprintf("Unable to parse running config: %s", err),
		)
		return
	}

	var inter cisconf.CiscoInterface
	for _, iface := range runningConfig.Interfaces {
		if iface.Parent.Identifier == data.ID.ValueString() {
			inter = iface
			break
		}
	}

	data = models.InterfaceSwitchFromCisconf(ctx, &inter)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *InterfaceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data models.InterfaceSwitchModel

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

	var inter cisconf.CiscoInterface
	for _, iface := range runningConfig.Interfaces {
		if iface.Parent.Identifier == data.ID.ValueString() {
			inter = iface
			break
		}
	}

	marshal, err := cisconf.Diff(inter, *models.InterfaceSwitchToCisconf(ctx, data))
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to diff interface",
			fmt.Sprintf("Unable to diff interface: %s", err),
		)
		return
	}
	lines := strings.Split(string(marshal), "\n")
	configs := []string{}
	for _, line := range lines {
		configs = append(configs, line)
	}
	tflog.Info(ctx, marshal)
	tflog.Info(ctx, fmt.Sprintf("%+v\n", inter))
	tflog.Info(ctx, fmt.Sprintf("%+v\n", *models.InterfaceSwitchToCisconf(ctx, data)))
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

	for _, iface := range runningConfig.Interfaces {
		if iface.Parent.Identifier == data.ID.ValueString() {
			inter = iface
			break
		}
	}

	data = models.InterfaceSwitchFromCisconf(ctx, &inter)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *InterfaceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data models.InterfaceModel

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

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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/sirikothe/gotextfsm"
	"strings"
	"terraform-provider-ios/internal/provider/models"
	"terraform-provider-ios/internal/provider/ntc"
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

	var vlan *cisconf.Vlan
	vlan = nil
	for _, v := range runningConfig.Vlans {
		if v.Id == int(data.Id.ValueInt32()) {
			vlan = &v
			break
		}
	}

	if vlan == nil {
		fsm, err := ntc.GetTextFSM("cisco_ios_show_vlan.textfsm")
		if err != nil {
			resp.Diagnostics.AddError(
				"Failed to get textfsm",
				fmt.Sprintf("Unable to get textfsm: %s", err),
			)
			return
		}
		result, err := r.client.Exec("sh vlan")
		if err != nil {
			resp.Diagnostics.AddError(
				"Failed to execute running config",
				fmt.Sprintf("Unable to execute running config: %s", err),
			)
			return
		}
		parser := gotextfsm.ParserOutput{}
		err = parser.ParseTextString(result, fsm, false)
		if err != nil {
			resp.Diagnostics.AddError(
				"Failed to parse result",
				fmt.Sprintf("Unable to parse result: %s", err),
			)
			return
		}
		for _, dic := range parser.Dict {
			if dic["VLAN_ID"] == fmt.Sprintf("%d", data.Id.ValueInt32()) {
				vlan = &cisconf.Vlan{
					Id:   int(data.Id.ValueInt32()),
					Name: dic["VLAN_NAME"].(string),
				}
				break
			}
		}
	}

	var marshal string
	if vlan == nil {
		marshal, err = cisconf.Marshal(models.VlanToCisconf(ctx, data))
	} else {
		marshal, err = cisconf.Diff(*vlan, models.VlanToCisconf(ctx, data))
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

	vlan = nil
	for _, v := range runningConfig.Vlans {
		tflog.Info(ctx, fmt.Sprintf("Vlan %d found", v.Id))
		if v.Id == int(data.Id.ValueInt32()) {
			vlan = &v
			break
		}
	}

	if vlan == nil {
		fsm, err := ntc.GetTextFSM("cisco_ios_show_vlan.textfsm")
		if err != nil {
			resp.Diagnostics.AddError(
				"Failed to get textfsm",
				fmt.Sprintf("Unable to get textfsm: %s", err),
			)
			return
		}
		result, err := r.client.Exec("sh vlan")
		if err != nil {
			resp.Diagnostics.AddError(
				"Failed to execute running config",
				fmt.Sprintf("Unable to execute running config: %s", err),
			)
			return
		}
		parser := gotextfsm.ParserOutput{}
		err = parser.ParseTextString(result, fsm, false)
		if err != nil {
			resp.Diagnostics.AddError(
				"Failed to parse result",
				fmt.Sprintf("Unable to parse result: %s", err),
			)
			return
		}
		for _, dic := range parser.Dict {
			if dic["VLAN_ID"] == fmt.Sprintf("%d", data.Id.ValueInt32()) {
				vlan = &cisconf.Vlan{
					Id:   int(data.Id.ValueInt32()),
					Name: dic["VLAN_NAME"].(string),
				}
				break
			}
		}
	}

	data = models.VlanFromCisconf(ctx, *vlan)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *VlanResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data models.VlanModel

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

	var vlan *cisconf.Vlan
	vlan = nil
	for _, v := range runningConfig.Vlans {
		if v.Id == int(data.Id.ValueInt32()) {
			vlan = &v
			break
		}
	}

	if vlan == nil {
		fsm, err := ntc.GetTextFSM("cisco_ios_show_vlan.textfsm")
		if err != nil {
			resp.Diagnostics.AddError(
				"Failed to get textfsm",
				fmt.Sprintf("Unable to get textfsm: %s", err),
			)
			return
		}
		result, err := r.client.Exec("sh vlan")
		if err != nil {
			resp.Diagnostics.AddError(
				"Failed to execute running config",
				fmt.Sprintf("Unable to execute running config: %s", err),
			)
			return
		}
		parser := gotextfsm.ParserOutput{}
		err = parser.ParseTextString(result, fsm, false)
		if err != nil {
			resp.Diagnostics.AddError(
				"Failed to parse result",
				fmt.Sprintf("Unable to parse result: %s", err),
			)
			return
		}
		for _, dic := range parser.Dict {
			if dic["VLAN_ID"] == fmt.Sprintf("%d", data.Id.ValueInt32()) {
				vlan = &cisconf.Vlan{
					Id:   int(data.Id.ValueInt32()),
					Name: dic["VLAN_NAME"].(string),
				}
				break
			}
		}
	}

	if vlan == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	data = models.VlanFromCisconf(ctx, *vlan)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *VlanResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data models.VlanModel

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

	var vlan *cisconf.Vlan
	vlan = nil
	for _, v := range runningConfig.Vlans {
		if v.Id == int(data.Id.ValueInt32()) {
			vlan = &v
			break
		}
	}

	if vlan == nil {
		fsm, err := ntc.GetTextFSM("cisco_ios_show_vlan.textfsm")
		if err != nil {
			resp.Diagnostics.AddError(
				"Failed to get textfsm",
				fmt.Sprintf("Unable to get textfsm: %s", err),
			)
			return
		}
		result, err := r.client.Exec("sh vlan")
		if err != nil {
			resp.Diagnostics.AddError(
				"Failed to execute running config",
				fmt.Sprintf("Unable to execute running config: %s", err),
			)
			return
		}
		parser := gotextfsm.ParserOutput{}
		err = parser.ParseTextString(result, fsm, false)
		if err != nil {
			resp.Diagnostics.AddError(
				"Failed to parse result",
				fmt.Sprintf("Unable to parse result: %s", err),
			)
			return
		}
		for _, dic := range parser.Dict {
			if dic["VLAN_ID"] == fmt.Sprintf("%d", data.Id.ValueInt32()) {
				vlan = &cisconf.Vlan{
					Id:   int(data.Id.ValueInt32()),
					Name: dic["VLAN_NAME"].(string),
				}
				break
			}
		}
	}

	var marshal string
	if vlan == nil {
		marshal, err = cisconf.Marshal(models.VlanToCisconf(ctx, data))
	} else {
		marshal, err = cisconf.Diff(*vlan, models.VlanToCisconf(ctx, data))
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

	vlan = nil
	for _, v := range runningConfig.Vlans {
		if v.Id == int(data.Id.ValueInt32()) {
			vlan = &v
			break
		}
	}

	if vlan == nil {
		fsm, err := ntc.GetTextFSM("cisco_ios_show_vlan.textfsm")
		if err != nil {
			resp.Diagnostics.AddError(
				"Failed to get textfsm",
				fmt.Sprintf("Unable to get textfsm: %s", err),
			)
			return
		}
		result, err := r.client.Exec("sh vlan")
		if err != nil {
			resp.Diagnostics.AddError(
				"Failed to execute running config",
				fmt.Sprintf("Unable to execute running config: %s", err),
			)
			return
		}
		parser := gotextfsm.ParserOutput{}
		err = parser.ParseTextString(result, fsm, false)
		if err != nil {
			resp.Diagnostics.AddError(
				"Failed to parse result",
				fmt.Sprintf("Unable to parse result: %s", err),
			)
			return
		}
		for _, dic := range parser.Dict {
			if dic["VLAN_ID"] == fmt.Sprintf("%d", data.Id.ValueInt32()) {
				vlan = &cisconf.Vlan{
					Id:   int(data.Id.ValueInt32()),
					Name: dic["VLAN_NAME"].(string),
				}
				break
			}
		}
	}

	data = models.VlanFromCisconf(ctx, *vlan)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
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

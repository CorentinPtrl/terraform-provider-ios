// Copyright (c) Corentin Pitrel
// SPDX-License-Identifier: MIT

package provider

import (
	"context"
	"fmt"
	"github.com/Letsu/cgnet"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sirikothe/gotextfsm"
	"slices"
	"strings"
)

var _ datasource.DataSource = &NtcDataSource{}

type NtcDataSourceModel struct {
	Data []map[string]interface{} `tfsdk:"data"`
}

func NewNtcDataSource(name string, fsm gotextfsm.TextFSM) datasource.DataSource {
	return &NtcDataSource{
		name: name,
		fsm:  fsm,
	}
}

type NtcDataSource struct {
	client *cgnet.Device
	name   string
	fsm    gotextfsm.TextFSM
}

func (d *NtcDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + d.name
}

func (d *NtcDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	fsmValues := map[string]schema.Attribute{}
	for k, field := range d.fsm.Values {
		if slices.Contains(field.Options, "List") {
			fsmValues[strings.ReplaceAll(strings.ToLower(k), "_", "")] = schema.ListAttribute{
				ElementType: types.StringType,
				Computed:    true,
			}
			continue
		}
		fsmValues[strings.ReplaceAll(strings.ToLower(k), "_", "")] = schema.StringAttribute{
			Computed: true,
		}
	}
	resp.Schema = schema.Schema{
		MarkdownDescription: "Data source for " + strings.ReplaceAll(d.name, "_", " "),

		Attributes: map[string]schema.Attribute{
			"data": schema.ListNestedAttribute{
				Computed:    true,
				Description: "Data source for " + strings.ReplaceAll(d.name, "_", " "),
				NestedObject: schema.NestedAttributeObject{
					Attributes: fsmValues,
				},
			},
		},
	}
}

func (d *NtcDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *NtcDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := d.client.Exec(strings.ReplaceAll(d.name, "_", " "))
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to execute running "+strings.ReplaceAll(d.name, "_", " "),
			fmt.Sprintf("Unable to execute running result: %s", err),
		)
		return
	}
	parser := gotextfsm.ParserOutput{}
	err = parser.ParseTextString(result, d.fsm, false)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to parse result",
			fmt.Sprintf("Unable to parse result: %s", err),
		)
		return
	}
	for index, dic := range parser.Dict {
		for k, v := range dic {
			if slices.Contains(d.fsm.Values[k].Options, "List") {
				list, diags := types.ListValueFrom(ctx, types.StringType, v.([]string))
				if diags.HasError() {
					resp.Diagnostics.AddError(
						"Failed to convert list value",
						fmt.Sprintf("Unable to convert list value: %s", diags),
					)
					return
				}
				resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("data").AtListIndex(index).AtName(strings.ReplaceAll(strings.ToLower(k), "_", "")), list)...)
				continue
			}
			resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("data").AtListIndex(index).AtName(strings.ReplaceAll(strings.ToLower(k), "_", "")), types.StringValue(v.(string)))...)
		}
	}
}

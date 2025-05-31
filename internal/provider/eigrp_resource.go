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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"terraform-provider-ios/internal/provider/models"
	"terraform-provider-ios/internal/utils"
)

var _ resource.Resource = &EigrpResource{}

func NewEigrpResource() resource.Resource {
	return &EigrpResource{}
}

type EigrpResource struct {
	client *cgnet.Device
}

func (r *EigrpResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_eigrp"
}

func (r *EigrpResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Static Route resource",

		Attributes: map[string]schema.Attribute{
			"networks": schema.ListAttribute{
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				Default:     listdefault.StaticValue(types.ListNull(types.StringType)),
			},
			"as_number": schema.Int64Attribute{
				Required: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *EigrpResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *EigrpResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data models.EigrpModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	eigrp, err := models.GetEigrpProcess(ctx, r.client, data.As.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to get EIGRP process",
			fmt.Sprintf("An error occurred while retrieving EIGRP process: %s", err),
		)
		return
	}

	var marshal string
	datacisco, err := models.EigrpToCisconf(ctx, data)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to convert EIGRP model to cisconf",
			fmt.Sprintf("Unable to convert EIGRP model: %s", err),
		)
		return
	}
	if eigrp == nil {
		marshal, err = cisconf.Marshal(datacisco)
	} else {
		var eigrpcisco cisconf.Eigrp
		eigrpcisco, err = models.EigrpToCisconf(ctx, data)
		if err != nil {
			resp.Diagnostics.AddError(
				"Failed to convert EIGRP model to cisconf",
				fmt.Sprintf("Unable to convert EIGRP model: %s", err),
			)
			return
		}
		marshal, err = cisconf.Diff(eigrpcisco, datacisco)
	}
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to marshal EIGRP configuration",
			fmt.Sprintf("Unable to marshal EIGRP configuration: %s", err),
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

	eigrp, err = models.GetEigrpProcess(ctx, r.client, data.As.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to get EIGRP process",
			fmt.Sprintf("An error occurred while retrieving EIGRP process: %s", err),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, eigrp)...)
}

func (r *EigrpResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data models.EigrpModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	eigrp, err := models.GetEigrpProcess(ctx, r.client, data.As.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to get EIGRP process",
			fmt.Sprintf("An error occurred while retrieving EIGRP process: %s", err),
		)
		return
	}

	if eigrp == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &eigrp)...)
}

func (r *EigrpResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data models.EigrpModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	eigrp, err := models.GetEigrpProcess(ctx, r.client, data.As.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to get EIGRP process",
			fmt.Sprintf("An error occurred while retrieving EIGRP process: %s", err),
		)
		return
	}

	var marshal string
	datacisco, err := models.EigrpToCisconf(ctx, data)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to convert EIGRP model to cisconf",
			fmt.Sprintf("Unable to convert EIGRP model: %s", err),
		)
		return
	}
	if eigrp == nil {
		marshal, err = cisconf.Marshal(datacisco)
	} else {
		var eigrpcisco cisconf.Eigrp
		eigrpcisco, err = models.EigrpToCisconf(ctx, data)
		if err != nil {
			resp.Diagnostics.AddError(
				"Failed to convert EIGRP model to cisconf",
				fmt.Sprintf("Unable to convert EIGRP model: %s", err),
			)
			return
		}
		marshal, err = cisconf.Diff(eigrpcisco, datacisco)
	}
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to marshal EIGRP configuration",
			fmt.Sprintf("Unable to marshal EIGRP configuration: %s", err),
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

	eigrp, err = models.GetEigrpProcess(ctx, r.client, data.As.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to get EIGRP process",
			fmt.Sprintf("An error occurred while retrieving EIGRP process: %s", err),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &eigrp)...)
}

func (r *EigrpResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data models.EigrpModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Configure([]string{"no router eigrp " + data.As.String()})
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to configure interface",
			fmt.Sprintf("Unable to configure interface: %s", err),
		)
		return
	}
}

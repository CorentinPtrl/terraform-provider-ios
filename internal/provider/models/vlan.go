// Copyright (c) Corentin Pitrel
// SPDX-License-Identifier: MIT

package models

import (
	"context"
	"github.com/CorentinPtrl/cisconf"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type VlanModel struct {
	Id   types.Int32  `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

func VlanToCisconf(ctx context.Context, data VlanModel) cisconf.Vlan {
	return cisconf.Vlan{
		Id:   int(data.Id.ValueInt32()),
		Name: data.Name.ValueString(),
	}
}

func VlanFromCisconf(ctx context.Context, vlan cisconf.Vlan) VlanModel {
	return VlanModel{
		Id:   types.Int32Value(int32(vlan.Id)),
		Name: types.StringValue(vlan.Name),
	}
}

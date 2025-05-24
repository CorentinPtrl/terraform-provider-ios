// Copyright (c) Corentin Pitrel
// SPDX-License-Identifier: MIT

package models

import (
	"context"
	"github.com/CorentinPtrl/cisconf"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type RoutesDataSourceModel struct {
	Routes []RouteModel `tfsdk:"routes"`
}

type RouteModel struct {
	Prefix  types.String `tfsdk:"prefix"`
	Mask    types.String `tfsdk:"mask"`
	NextHop types.String `tfsdk:"next_hop"`
}

func RouteToCisconf(ctx context.Context, route RouteModel) cisconf.Route {
	return cisconf.Route{
		Prefix:    route.Prefix.ValueString(),
		Mask:      route.Mask.ValueString(),
		IpAddress: route.NextHop.ValueString(),
	}
}

func RouteFromCisconf(ctx context.Context, route cisconf.Route) RouteModel {
	return RouteModel{
		Prefix:  types.StringValue(route.Prefix),
		Mask:    types.StringValue(route.Mask),
		NextHop: types.StringValue(route.IpAddress),
	}
}

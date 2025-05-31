// Copyright (c) Corentin Pitrel
// SPDX-License-Identifier: MIT

package models

import (
	"fmt"
	"github.com/CorentinPtrl/cgnet"
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

func RouteToCisconf(route RouteModel) cisconf.Route {
	return cisconf.Route{
		Prefix:    route.Prefix.ValueString(),
		Mask:      route.Mask.ValueString(),
		IpAddress: route.NextHop.ValueString(),
	}
}

func RouteFromCisconf(route cisconf.Route) RouteModel {
	return RouteModel{
		Prefix:  types.StringValue(route.Prefix),
		Mask:    types.StringValue(route.Mask),
		NextHop: types.StringValue(route.IpAddress),
	}
}

func GetRoutes(device *cgnet.Device) ([]RouteModel, error) {
	config, err := device.Exec("sh running-config")
	if err != nil {
		return nil, fmt.Errorf("failed to execute running config: %w", err)
	}
	runningConfig := cisconf.Config{}
	err = cisconf.Unmarshal(config, &runningConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal running config: %w", err)
	}

	result := []RouteModel{}
	for _, v := range runningConfig.Routes {
		result = append(result, RouteFromCisconf(v))
	}
	return result, nil
}

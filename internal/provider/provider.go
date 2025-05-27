// Copyright (c) Corentin Pitrel
// SPDX-License-Identifier: MIT

package provider

import (
	"context"
	"github.com/Letsu/cgnet"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"os"
	"strings"
	"terraform-provider-ios/internal/provider/ntc"
)

// Ensure CiscoIosProvider satisfies various provider interfaces.
var _ provider.Provider = &CiscoIosProvider{}
var _ provider.ProviderWithFunctions = &CiscoIosProvider{}
var _ provider.ProviderWithEphemeralResources = &CiscoIosProvider{}

// CiscoIosProvider defines the provider implementation.
type CiscoIosProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// CiscoIosProviderModel describes the provider data model.
type CiscoIosProviderModel struct {
	Host     types.String `tfsdk:"host"`
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
}

func (p *CiscoIosProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "ios"
	resp.Version = p.version
}

func (p *CiscoIosProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"host": schema.StringAttribute{
				Optional: true,
			},
			"username": schema.StringAttribute{
				Optional: true,
			},
			"password": schema.StringAttribute{
				Optional:  true,
				Sensitive: true,
			},
		},
	}
}

func (p *CiscoIosProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config CiscoIosProviderModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if config.Host.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("host"),
			"Unknown Cisco IOS Host",
			"The provider cannot create the Cisco IOS ssh client as there is an unknown configuration value for the Cisco IOS host. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the IOS_HOST environment variable.",
		)
	}

	if config.Username.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("username"),
			"Unknown Cisco IOS Username",
			"The provider cannot create the Cisco IOS client as there is an unknown configuration value for the Cisco IOS username. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the IOS_USERNAME environment variable.",
		)
	}

	if config.Password.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("password"),
			"Unknown Cisco IOS Password",
			"The provider cannot create the Cisco IOS client as there is an unknown configuration value for the Cisco IOS password. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the IOS_PASSWORD environment variable.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	host := os.Getenv("IOS_HOST")
	username := os.Getenv("IOS_USERNAME")
	password := os.Getenv("IOS_PASSWORD")

	if !config.Host.IsNull() {
		host = config.Host.ValueString()
	}

	if !config.Username.IsNull() {
		username = config.Username.ValueString()
	}

	if !config.Password.IsNull() {
		password = config.Password.ValueString()
	}

	if host == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("host"),
			"Missing Cisco IOS Host",
			"The provider cannot create the Cisco IOS client as there is a missing or empty value for the Cisco IOS host. "+
				"Set the host value in the configuration or use the IOS_HOST environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if username == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("username"),
			"Missing Cisco IOS Username",
			"The provider cannot create the Cisco IOS client as there is a missing or empty value for the Cisco IOS username. "+
				"Set the username value in the configuration or use the IOS_USERNAME environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if password == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("password"),
			"Missing Cisco IOS Password",
			"The provider cannot create the Cisco IOS client as there is a missing or empty value for the Cisco IOS password. "+
				"Set the password value in the configuration or use the IOS_PASSWORD environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	session := cgnet.Device{
		Ip:       host,
		Port:     "22",
		Username: username,
		Password: password,
		ConnType: cgnet.SSH,
	}
	err := session.Open()
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Cisco IOS Client",
			"An unexpected error occurred when creating the Cisco IOS client. "+
				"If the error is not clear, please contact the provider developers.\n\n"+
				"Cisco IOS Error: "+err.Error(),
		)
		return
	}

	resp.DataSourceData = &session
	resp.ResourceData = &session
}

func (p *CiscoIosProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewVlanResource,
		NewInterfaceSwitchResource,
		NewInterfaceEthernetResource,
		NewStaticRouteResource,
		NewEigrpResource,
	}
}

func (p *CiscoIosProvider) EphemeralResources(ctx context.Context) []func() ephemeral.EphemeralResource {
	return []func() ephemeral.EphemeralResource{}
}

func (p *CiscoIosProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	datasources := []func() datasource.DataSource{
		NewVlanDataSource,
		NewVlansDataSource,
		NewInterfacesDataSource,
		NewStaticRoutesDataSource,
	}
	temps, err := ntc.GetTemplateNames()
	if err != nil {
		panic(err)
	}
	for _, template := range temps {
		textfsm, err := ntc.GetTextFSM(template)
		if err != nil {
			continue
		}
		datasources = append(datasources, func() datasource.DataSource {
			return NewNtcDataSource(strings.ReplaceAll(strings.ReplaceAll(template, ".textfsm", ""), "cisco_ios_", ""), textfsm)
		})
	}
	return datasources
}

func (p *CiscoIosProvider) Functions(ctx context.Context) []func() function.Function {
	return []func() function.Function{}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &CiscoIosProvider{
			version: version,
		}
	}
}

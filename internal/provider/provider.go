package provider

import (
	"context"
	"fmt"
	"os"
	"terraform-provider-gsolaceclustermgr/internal/missioncontrol"

	// "missioncontrol"

	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// clusterManagerProviderModel maps provider schema data to a Go type.
type clusterManagerProviderModel struct {
	Host        types.String `tfsdk:"host"`
	BearerToken types.String `tfsdk:"bearer_token"`
}

// Ensure the implementation satisfies the expected interfaces.
var (
	_ provider.Provider = &clusterManagerProvider{}
)

// New is a helper function to simplify provider server and testing implementation.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &clusterManagerProvider{
			version: version,
		}
	}
}

// clusterManagerProvider is the provider implementation.
type clusterManagerProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

type MCClientHolder struct {
	Client      *missioncontrol.ClientWithResponses
	BearerToken string
}

// Metadata returns the provider type name.
func (p *clusterManagerProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "gsolaceclustermgr"
	resp.Version = p.version
}

// Schema defines the provider-level schema for configuration data.
func (p *clusterManagerProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"host": schema.StringAttribute{
				Required: true,
			},
			"bearer_token": schema.StringAttribute{
				Required:  true,
				Sensitive: true,
			},
		},
	}
}

// Configure prepares a Solace MissionControl API client for data sources and resources.
func (p *clusterManagerProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	tflog.Info(ctx, "Configuring ClusterManager Provider")

	// Retrieve provider data from configuration
	var config clusterManagerProviderModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// If practitioner provided a configuration value for any of the
	// attributes, it must be a known value.

	if config.Host.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("host"),
			"Unknown HashiCups API Host",
			"The provider cannot create the MissionControl API client as there is an unknown configuration value for the HashiCups API host. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the MISSIONCONTROL_HOST environment variable.",
		)
	}

	if config.BearerToken.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("password"),
			"Unknown MissionControl API Token",
			"The provider cannot create the MissionControl API client as there is an unknown configuration value for the MissionControl API token. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the MISSIONCONTROL_TOKEN environment variable.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Default values to environment variables, but override
	// with Terraform configuration value if set.

	host := os.Getenv("MISSIONCONTROL_HOST")
	bearerToken := os.Getenv("MISSIONCONTROL_TOKEN")

	if !config.Host.IsNull() {
		host = config.Host.ValueString()
	}

	if !config.BearerToken.IsNull() {
		bearerToken = config.BearerToken.ValueString()
	}

	// If any of the expected configurations are missing, return
	// errors with provider-specific guidance.

	if host == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("host"),
			"Missing MissionControl API Host",
			"The provider cannot create the MissionControl API client as there is a missing or empty value for the MissionControl API host. "+
				"Set the host value in the configuration or use the MISSIONCONTROL_HOST environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if bearerToken == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("bearerToken"),
			"Missing MissionControl API PasTokensword",
			"The provider cannot create the MissionControl API client as there is a missing or empty value for the MissionControl API token. "+
				"Set the password value in the configuration or use the MISSIONCONTROL_TOKEN environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	ctx = tflog.SetField(ctx, "missioncontrol_host", host)
	ctx = tflog.SetField(ctx, "missioncontrol_token", bearerToken)
	ctx = tflog.MaskFieldValuesWithFieldKeys(ctx, "missioncontrol_token")

	tflog.Info(ctx, fmt.Sprintf("Creating MissionControl client using %s", host))

	// Create a new  client using the configuration values
	// custom HTTP client
	hc := http.Client{}

	// TODO how to treat token...
	client, err := missioncontrol.NewClientWithResponses(host, missioncontrol.WithHTTPClient(&hc))
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create MissionControl API Client",
			"An unexpected error occurred when creating the HashiCups API client. "+
				"If the error is not clear, please contact the provider developers.\n\n"+
				"HashiCups Client Error: "+err.Error(),
		)
		return
	}

	// Make the HashiCups client available during DataSource and Resource
	// type Configure methods.
	resp.DataSourceData = client
	resp.ResourceData = MCClientHolder{client, bearerToken}

	tflog.Info(ctx, "Configured MissionControl client", map[string]any{"success": true})
}

// DataSources defines the data sources implemented in the provider.
func (p *clusterManagerProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return nil
}

// Resources defines the resources implemented in the provider.
func (p *clusterManagerProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewBrokerResource,
	}
}

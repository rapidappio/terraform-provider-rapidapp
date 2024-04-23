package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/rapidappio/rapidapp-go"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ provider.Provider = &rapidappProvider{}
)

type rapidappProviderModel struct {
	ApiKey types.String `tfsdk:"api_key"`
}

// New is a helper function to simplify provider server and testing implementation.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &rapidappProvider{
			version: version,
		}
	}
}

// rapidappProvider is the provider implementation.
type rapidappProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// Metadata returns the provider type name.
func (p *rapidappProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "rapidapp"
	resp.Version = p.version
}

// Schema defines the provider-level schema for configuration data.
func (p *rapidappProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"api_key": schema.StringAttribute{
				Required:  true,
				Sensitive: true,
			},
		},
	}
}

// Configure prepares a Rapidapp API client for data sources and resources.
func (p *rapidappProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	// Retrieve provider data from configuration
	var config rapidappProviderModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if config.ApiKey.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("api_key"),
			"Unknown Api Key value",
			"The provider cannot create the Rapidapp client as there is an unknown configuration value for the Api Key.",
		)
	}
	if resp.Diagnostics.HasError() {
		return
	}

	apiKey := os.Getenv("RAPIDAPP_API_KEY")

	if !config.ApiKey.IsNull() {
		apiKey = config.ApiKey.ValueString()
	}

	if apiKey == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("api_key"),
			"Missing Api Key",
			"The provider cannot create the Rapidapp client as there is a missing or empty value for the Api Key. ",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}
	// Create Rapidapp client
	client := rapidapp.NewClient(apiKey)

	resp.DataSourceData = client
	resp.ResourceData = client
}

// DataSources defines the data sources implemented in the provider.
func (p *rapidappProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewPostgresDatabaseDataSource,
	}
}

// Resources defines the resources implemented in the provider.
func (p *rapidappProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewPostgresDatabaseResource,
	}
}

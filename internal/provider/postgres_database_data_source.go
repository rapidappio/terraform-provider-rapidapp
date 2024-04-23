package provider

import (
	"context"
	"fmt"
	"github.com/rapidappio/rapidapp-go"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	tftypes "github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = &postgresDatabaseDataSource{}
	_ datasource.DataSourceWithConfigure = &postgresDatabaseDataSource{}
)

// NewPostgresDatabaseDataSource is a helper function to simplify the provider implementation.
func NewPostgresDatabaseDataSource() datasource.DataSource {
	return &postgresDatabaseDataSource{}
}

// postgresDatabaseDataSource is the data source implementation.
type postgresDatabaseDataSource struct {
	client *rapidapp.Client
}

// postgresDatabaseModel maps postgres schema data.
type postgresDatabaseModel struct {
	ID     tftypes.String `tfsdk:"id"`
	Name   tftypes.String `tfsdk:"name"`
	Status tftypes.String `tfsdk:"status"`
}

func (d *postgresDatabaseDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_postgres_database"
}

func (d *postgresDatabaseDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required: true,
			},
			"name": schema.StringAttribute{
				Optional: true,
			},
			"status": schema.StringAttribute{
				Optional: true,
			},
		},
	}
}

func (d *postgresDatabaseDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state postgresDatabaseModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	db, err := d.client.GetPostgresDatabase(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Postgres databases data",
			err.Error(),
		)
		return
	}

	state = postgresDatabaseModel{
		ID:     types.StringValue(db.Id),
		Name:   types.StringValue(db.Name),
		Status: types.StringValue(db.Status),
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (d *postgresDatabaseDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*rapidapp.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *rapidapp.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.client = client
}

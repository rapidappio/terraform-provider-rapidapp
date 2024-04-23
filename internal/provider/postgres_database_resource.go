package provider

import (
	"context"
	"fmt"
	"github.com/rapidappio/rapidapp-go"
	"github.com/rapidappio/rapidapp-go/pkg/proto/postgres"
	"log"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	tftypes "github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource = &postgresDatabaseResource{}
)

// NewPostgresDatabaseResource is a helper function to simplify the provider implementation.
func NewPostgresDatabaseResource() resource.Resource {
	return &postgresDatabaseResource{}
}

// postgresDatabaseResource is the resource implementation.
type postgresDatabaseResource struct {
	client *rapidapp.Client
}

// postgresDatabaseResourceModel maps the resource schema data.
type postgresDatabaseResourceModel struct {
	ID     tftypes.String `tfsdk:"id"`
	Name   tftypes.String `tfsdk:"name"`
	Status tftypes.String `tfsdk:"status"`
}

// Metadata returns the resource type name.
func (r *postgresDatabaseResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_postgres_database"
}

// Schema defines the schema for the resource.
func (r *postgresDatabaseResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"name": schema.StringAttribute{
				Required: true,
			},
			"status": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

// Create a new resource.
func (r *postgresDatabaseResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {

	var plan postgresDatabaseResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := r.client.CreatePostgresDatabase(plan.Name.String())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating postgres database",
			"Could not create postgres database, unexpected error: "+err.Error(),
		)
		return
	}

	d, err := r.waitForPostgresDatabaseReady(ctx, id)

	if err != nil {
		resp.Diagnostics.AddError(
			"Error waiting for postgres database",
			"Could not wait for postgres database to be ready, unexpected error: "+err.Error(),
		)
		return
	}

	plan.ID = types.StringValue(d.Id)
	plan.Status = types.StringValue(d.Status)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read resource information.
func (r *postgresDatabaseResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state postgresDatabaseResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.waitForPostgresDatabaseReady(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error waiting for postgres database",
			"Could not wait for postgres database to be ready, unexpected error: "+err.Error(),
		)
		return
	}

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *postgresDatabaseResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	log.Fatal("Update not implemented")
}

func (r *postgresDatabaseResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from state
	var state postgresDatabaseResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeletePostgresDatabase(state.ID.ValueString())

	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting postgres database",
			"Could not delete postgres database, unexpected error: "+err.Error(),
		)
		return

	}
	err = r.waitForPostgresDatabaseDeleted(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error waiting for postgres database to be deleted",
			"Could not wait for postgres database to be deleted, unexpected error: "+err.Error(),
		)
		return
	}

	state = postgresDatabaseResourceModel{}
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

}

// Configure adds the provider configured client to the resource.
func (r *postgresDatabaseResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

	r.client = client
}

func (r *postgresDatabaseResource) waitForPostgresDatabaseReady(ctx context.Context, id string) (*postgres.Postgres, error) {
	// Define a timeout for waiting (adjust as needed)
	deadline, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	for {
		database, err := r.client.GetPostgresDatabase(id)
		if err == nil && database.Status == "running" {
			return database, nil
		}
		select {
		case <-deadline.Done():
			return nil, fmt.Errorf("timeout waiting for database %s to be ready", id)
		case <-time.After(time.Second):
			// Retry after a brief interval
		}
	}
}

func (r *postgresDatabaseResource) waitForPostgresDatabaseDeleted(ctx context.Context, id string) error {
	// Define a timeout for waiting (adjust as needed)
	deadline, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	for {
		_, err := r.client.GetPostgresDatabase(id)
		if err != nil && strings.Contains(err.Error(), "not found") {
			return nil
		}
		select {
		case <-deadline.Done():
			return fmt.Errorf("timeout waiting for database %s to be deleted", id)
		case <-time.After(time.Second):
			// Retry after a brief interval
		}
	}
}

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/gustavebbers/terraform-provider-confluence-dc/internal/client"
)

var _ datasource.DataSource = &spaceDataSource{}
var _ datasource.DataSourceWithConfigure = &spaceDataSource{}

// NewSpaceDataSource instantiates the confluencedc_space data source.
func NewSpaceDataSource() datasource.DataSource {
	return &spaceDataSource{}
}

type spaceDataSource struct {
	client *client.Client
}

type spaceDataSourceModel struct {
	Key         types.String `tfsdk:"key"`
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Type        types.String `tfsdk:"type"`
	Description types.String `tfsdk:"description"`
}

func (d *spaceDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_space"
}

func (d *spaceDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves information about an existing Confluence space by its key.",
		Attributes: map[string]schema.Attribute{
			"key": schema.StringAttribute{
				Required:    true,
				Description: "The unique key of the space, e.g. \"ENG\".",
			},
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The internal numeric ID of the space.",
			},
			"name": schema.StringAttribute{
				Computed:    true,
				Description: "The display name of the space.",
			},
			"type": schema.StringAttribute{
				Computed:    true,
				Description: "The type of the space, e.g. \"global\" or \"personal\".",
			},
			"description": schema.StringAttribute{
				Computed:    true,
				Description: "The plain-text space description.",
			},
		},
	}
}

func (d *spaceDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}
	d.client = c
}

func (d *spaceDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data spaceDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	space, err := d.client.GetSpace(ctx, data.Key.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Confluence Space",
			fmt.Sprintf("Could not read space %q: %s", data.Key.ValueString(), err),
		)
		return
	}

	data.ID = types.StringValue(fmt.Sprintf("%d", space.ID))
	data.Name = types.StringValue(space.Name)
	data.Type = types.StringValue(space.Type)
	data.Description = types.StringValue(space.Description.Plain.Value)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

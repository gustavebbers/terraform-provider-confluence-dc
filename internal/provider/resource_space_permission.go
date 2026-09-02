package provider

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/gustavebbers/terraform-provider-confluence-dc/internal/client"
)

var _ resource.Resource = &spacePermissionResource{}
var _ resource.ResourceWithConfigure = &spacePermissionResource{}
var _ resource.ResourceWithImportState = &spacePermissionResource{}

// NewSpacePermissionResource instantiates the confluencedc_space_permission resource.
func NewSpacePermissionResource() resource.Resource {
	return &spacePermissionResource{}
}

type spacePermissionResource struct {
	client *client.Client
}

type spacePermissionResourceModel struct {
	ID              types.String `tfsdk:"id"`
	SpaceKey        types.String `tfsdk:"space_key"`
	GroupName       types.String `tfsdk:"group_name"`
	OperationKey    types.String `tfsdk:"operation_key"`
	OperationTarget types.String `tfsdk:"operation_target"`
}

func (r *spacePermissionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_space_permission"
}

func (r *spacePermissionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Grants a permission on a Confluence space to a group. Requires Confluence Data Center " +
			"9.1 or later, the first release to expose space permission management through the REST API.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Composite identifier in the form \"<space_key>/<permission_id>\".",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"space_key": schema.StringAttribute{
				Required:    true,
				Description: "Key of the space the permission applies to.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"group_name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the group the permission is granted to.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"operation_key": schema.StringAttribute{
				Required: true,
				Description: "The operation being granted, e.g. \"read\", \"create\", \"delete\", \"export\", " +
					"\"administer\", \"restrict_content\", or \"archive\". See the Confluence REST API " +
					"documentation for valid operation_key/operation_target combinations.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"operation_target": schema.StringAttribute{
				Required:    true,
				Description: "The content type the operation applies to, e.g. \"space\", \"page\", \"blogpost\", \"comment\", or \"attachment\".",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *spacePermissionResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}
	r.client = c
}

func (r *spacePermissionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan spacePermissionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	perm, err := r.client.AddSpacePermission(ctx, plan.SpaceKey.ValueString(),
		client.PermissionSubject{Type: "group", Identifier: plan.GroupName.ValueString()},
		client.PermissionOperation{Key: plan.OperationKey.ValueString(), Target: plan.OperationTarget.ValueString()},
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Confluence Space Permission",
			fmt.Sprintf("Could not grant %s/%s on space %q to group %q: %s",
				plan.OperationKey.ValueString(), plan.OperationTarget.ValueString(),
				plan.SpaceKey.ValueString(), plan.GroupName.ValueString(), err),
		)
		return
	}

	plan.ID = types.StringValue(composePermissionID(plan.SpaceKey.ValueString(), perm.ID))

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *spacePermissionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state spacePermissionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	spaceKey, permID, err := parsePermissionID(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid Resource ID", err.Error())
		return
	}

	perm, err := r.client.GetSpacePermission(ctx, spaceKey, permID)
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Unable to Read Confluence Space Permission",
			fmt.Sprintf("Could not read permission %d on space %q: %s", permID, spaceKey, err),
		)
		return
	}

	state.ID = types.StringValue(composePermissionID(spaceKey, perm.ID))
	state.SpaceKey = types.StringValue(spaceKey)
	state.GroupName = types.StringValue(perm.Subject.Identifier)
	state.OperationKey = types.StringValue(perm.Operation.Key)
	state.OperationTarget = types.StringValue(perm.Operation.Target)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *spacePermissionResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Update Not Supported",
		"confluencedc_space_permission does not support in-place updates; all attribute changes force replacement.",
	)
}

func (r *spacePermissionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state spacePermissionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	spaceKey, permID, err := parsePermissionID(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid Resource ID", err.Error())
		return
	}

	if err := r.client.RemoveSpacePermission(ctx, spaceKey, permID); err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError(
			"Unable to Delete Confluence Space Permission",
			fmt.Sprintf("Could not delete permission %d on space %q: %s", permID, spaceKey, err),
		)
	}
}

func (r *spacePermissionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func composePermissionID(spaceKey string, permissionID int64) string {
	return fmt.Sprintf("%s/%d", spaceKey, permissionID)
}

func parsePermissionID(id string) (spaceKey string, permissionID int64, err error) {
	parts := strings.SplitN(id, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", 0, fmt.Errorf("expected an ID in the form \"<space_key>/<permission_id>\", got: %q", id)
	}

	permissionID, err = strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return "", 0, fmt.Errorf("expected a numeric permission ID in %q: %w", id, err)
	}

	return parts[0], permissionID, nil
}

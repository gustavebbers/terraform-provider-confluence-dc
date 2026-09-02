package provider

import (
	"context"
	"fmt"
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
		Description: "Grants a permission on a Confluence space to a group. Confluence Data Center's REST API " +
			"only supports reading space permissions, not granting or revoking them, so this resource performs " +
			"writes through Confluence's legacy JSON-RPC API (confluenceservice-v2) instead; that API is " +
			"deprecated by Atlassian but still present and functional as of Confluence Data Center 9.2. It must " +
			"remain enabled on the target instance for this resource to work.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				Description: "Composite identifier in the form " +
					"\"<space_key>/<group_name>/<operation_key>/<operation_target>\".",
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
				Description: "The operation being granted. One of: \"read\", \"create\", \"delete\", " +
					"\"delete_own\", \"delete_mail\", \"export\", \"restrict\", \"administer\". Must be paired " +
					"with a valid operation_target; see the resource description for the full list of valid pairs.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"operation_target": schema.StringAttribute{
				Required: true,
				Description: "The content type the operation applies to. One of: \"space\", \"page\", " +
					"\"blogpost\", \"comment\", \"attachment\". Valid operation_key/operation_target pairs: " +
					"read/space, create/page, delete/page, create/blogpost, delete/blogpost, create/comment, " +
					"delete/comment, create/attachment, delete/attachment, delete_own/space, delete_mail/space, " +
					"export/space, restrict/space, administer/space.",
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

	spaceKey := plan.SpaceKey.ValueString()
	groupName := plan.GroupName.ValueString()
	operationKey := plan.OperationKey.ValueString()
	operationTarget := plan.OperationTarget.ValueString()

	_, err := r.client.AddSpacePermission(ctx, spaceKey,
		client.PermissionSubject{Type: "group", Identifier: groupName},
		client.PermissionOperation{Key: operationKey, Target: operationTarget},
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Confluence Space Permission",
			fmt.Sprintf("Could not grant %s/%s on space %q to group %q: %s",
				operationKey, operationTarget, spaceKey, groupName, err),
		)
		return
	}

	plan.ID = types.StringValue(composePermissionID(spaceKey, groupName, operationKey, operationTarget))

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *spacePermissionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state spacePermissionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	spaceKey, groupName, operationKey, operationTarget, err := parsePermissionID(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid Resource ID", err.Error())
		return
	}

	_, err = r.client.GetSpacePermission(ctx, spaceKey,
		client.PermissionSubject{Type: "group", Identifier: groupName},
		client.PermissionOperation{Key: operationKey, Target: operationTarget},
	)
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Unable to Read Confluence Space Permission",
			fmt.Sprintf("Could not read %s/%s permission for group %q on space %q: %s",
				operationKey, operationTarget, groupName, spaceKey, err),
		)
		return
	}

	state.ID = types.StringValue(composePermissionID(spaceKey, groupName, operationKey, operationTarget))
	state.SpaceKey = types.StringValue(spaceKey)
	state.GroupName = types.StringValue(groupName)
	state.OperationKey = types.StringValue(operationKey)
	state.OperationTarget = types.StringValue(operationTarget)

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

	spaceKey, groupName, operationKey, operationTarget, err := parsePermissionID(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid Resource ID", err.Error())
		return
	}

	err = r.client.RemoveSpacePermission(ctx, spaceKey,
		client.PermissionSubject{Type: "group", Identifier: groupName},
		client.PermissionOperation{Key: operationKey, Target: operationTarget},
	)
	if err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError(
			"Unable to Delete Confluence Space Permission",
			fmt.Sprintf("Could not revoke %s/%s permission for group %q on space %q: %s",
				operationKey, operationTarget, groupName, spaceKey, err),
		)
	}
}

func (r *spacePermissionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// composePermissionID builds the resource's composite ID. Space keys and
// group names containing "/" are not supported for import (parsePermissionID
// cannot unambiguously split them back apart); this is documented as a
// limitation rather than worked around, since Confluence space keys never
// contain "/" and group names doing so are exceedingly rare.
func composePermissionID(spaceKey, groupName, operationKey, operationTarget string) string {
	return strings.Join([]string{spaceKey, groupName, operationKey, operationTarget}, "/")
}

func parsePermissionID(id string) (spaceKey, groupName, operationKey, operationTarget string, err error) {
	parts := strings.Split(id, "/")
	if len(parts) != 4 {
		return "", "", "", "", fmt.Errorf(
			"expected an ID in the form \"<space_key>/<group_name>/<operation_key>/<operation_target>\", got: %q", id)
	}
	for _, p := range parts {
		if p == "" {
			return "", "", "", "", fmt.Errorf(
				"expected an ID in the form \"<space_key>/<group_name>/<operation_key>/<operation_target>\" with no empty segments, got: %q", id)
		}
	}
	return parts[0], parts[1], parts[2], parts[3], nil
}

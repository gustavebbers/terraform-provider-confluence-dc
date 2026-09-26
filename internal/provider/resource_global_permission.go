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

var _ resource.Resource = &globalPermissionResource{}
var _ resource.ResourceWithConfigure = &globalPermissionResource{}
var _ resource.ResourceWithImportState = &globalPermissionResource{}

// NewGlobalPermissionResource instantiates the confluencedc_global_permission resource.
func NewGlobalPermissionResource() resource.Resource {
	return &globalPermissionResource{}
}

type globalPermissionResource struct {
	client *client.Client
}

type globalPermissionResourceModel struct {
	ID              types.String `tfsdk:"id"`
	GroupName       types.String `tfsdk:"group_name"`
	OperationKey    types.String `tfsdk:"operation_key"`
	OperationTarget types.String `tfsdk:"operation_target"`
}

func (r *globalPermissionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_global_permission"
}

func (r *globalPermissionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Grants a global (instance-wide) permission to a group, via " +
			"PUT /rest/api/permissions/group/{groupName}/grant (and .../revoke on destroy). Use this for " +
			"instance-wide access such as letting a group log in and use Confluence at all (use/application), " +
			"or granting system administrator rights (administer/system) - as opposed to " +
			"confluencedc_space_permission, which scopes a permission to one space. There is no legacy " +
			"JSON-RPC fallback for this resource: global permission management was never exposed by " +
			"Confluence's legacy remote API, so it requires the REST endpoint above to be present on the " +
			"target instance.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				Description: "Composite identifier in the form " +
					"\"<group_name>/<operation_key>/<operation_target>\".",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
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
				Description: "The operation being granted. One of: \"use\", \"administer\", \"create\". Must " +
					"be paired with a valid operation_target; see the resource description for the full list " +
					"of valid pairs.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"operation_target": schema.StringAttribute{
				Required: true,
				Description: "What the operation applies to. One of: \"application\", \"system\", " +
					"\"personal_space\", \"space\". Valid operation_key/operation_target pairs: use/application " +
					"(standard login/use access), administer/application, administer/system (system " +
					"administrator), create/personal_space, create/space.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *globalPermissionResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *globalPermissionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan globalPermissionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	groupName := plan.GroupName.ValueString()
	operationKey := plan.OperationKey.ValueString()
	operationTarget := plan.OperationTarget.ValueString()

	err := r.client.AddGlobalPermission(ctx, groupName,
		client.GlobalPermissionOperation{Key: operationKey, Target: operationTarget},
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Confluence Global Permission",
			fmt.Sprintf("Could not grant %s/%s to group %q: %s", operationKey, operationTarget, groupName, err),
		)
		return
	}

	plan.ID = types.StringValue(composeGlobalPermissionID(groupName, operationKey, operationTarget))

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *globalPermissionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state globalPermissionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	groupName, operationKey, operationTarget, err := parseGlobalPermissionID(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid Resource ID", err.Error())
		return
	}

	_, err = r.client.GetGlobalPermission(ctx, groupName,
		client.GlobalPermissionOperation{Key: operationKey, Target: operationTarget},
	)
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Unable to Read Confluence Global Permission",
			fmt.Sprintf("Could not read %s/%s permission for group %q: %s", operationKey, operationTarget, groupName, err),
		)
		return
	}

	state.ID = types.StringValue(composeGlobalPermissionID(groupName, operationKey, operationTarget))
	state.GroupName = types.StringValue(groupName)
	state.OperationKey = types.StringValue(operationKey)
	state.OperationTarget = types.StringValue(operationTarget)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *globalPermissionResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Update Not Supported",
		"confluencedc_global_permission does not support in-place updates; all attribute changes force replacement.",
	)
}

func (r *globalPermissionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state globalPermissionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	groupName, operationKey, operationTarget, err := parseGlobalPermissionID(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid Resource ID", err.Error())
		return
	}

	err = r.client.RemoveGlobalPermission(ctx, groupName,
		client.GlobalPermissionOperation{Key: operationKey, Target: operationTarget},
	)
	if err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError(
			"Unable to Delete Confluence Global Permission",
			fmt.Sprintf("Could not revoke %s/%s permission for group %q: %s", operationKey, operationTarget, groupName, err),
		)
	}
}

func (r *globalPermissionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// composeGlobalPermissionID builds the resource's composite ID. Group names
// containing "/" are not supported for import (parseGlobalPermissionID
// cannot unambiguously split them back apart); this is documented as a
// limitation rather than worked around, since group names doing so are
// exceedingly rare.
func composeGlobalPermissionID(groupName, operationKey, operationTarget string) string {
	return strings.Join([]string{groupName, operationKey, operationTarget}, "/")
}

func parseGlobalPermissionID(id string) (groupName, operationKey, operationTarget string, err error) {
	parts := strings.Split(id, "/")
	if len(parts) != 3 {
		return "", "", "", fmt.Errorf(
			"expected an ID in the form \"<group_name>/<operation_key>/<operation_target>\", got: %q", id)
	}
	for _, p := range parts {
		if p == "" {
			return "", "", "", fmt.Errorf(
				"expected an ID in the form \"<group_name>/<operation_key>/<operation_target>\" with no empty segments, got: %q", id)
		}
	}
	return parts[0], parts[1], parts[2], nil
}

package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/gustavebbers/terraform-provider-confluence-dc/internal/client"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ provider.Provider = &confluenceDCProvider{}
)

// New returns a function that instantiates the provider, for use with
// providerserver.Serve.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &confluenceDCProvider{version: version}
	}
}

type confluenceDCProvider struct {
	version string
}

// confluenceDCProviderModel maps the provider configuration block.
type confluenceDCProviderModel struct {
	Host          types.String `tfsdk:"host"`
	Token         types.String `tfsdk:"token"`
	Username      types.String `tfsdk:"username"`
	Password      types.String `tfsdk:"password"`
	SkipTLSVerify types.Bool   `tfsdk:"skip_tls_verify"`
}

func (p *confluenceDCProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "confluencedc"
	resp.Version = p.version
}

func (p *confluenceDCProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Interacts with a Confluence Data Center (self-managed) instance to manage groups and " +
			"space permissions, and to read space information. Confluence Data Center's REST API does not " +
			"support creating/deleting groups or granting/revoking space permissions, so confluencedc_group " +
			"and confluencedc_space_permission perform writes through Confluence's legacy JSON-RPC API " +
			"(confluenceservice-v2) instead; that API is deprecated by Atlassian but present and functional on " +
			"current Data Center releases. It must remain enabled on the target instance.",
		Attributes: map[string]schema.Attribute{
			"host": schema.StringAttribute{
				Optional: true,
				Description: "Base URL of the Confluence Data Center instance, e.g. https://confluence.example.com. " +
					"May also be set via the CONFLUENCE_HOST environment variable.",
			},
			"token": schema.StringAttribute{
				Optional:  true,
				Sensitive: true,
				Description: "Personal Access Token used to authenticate via HTTP Bearer auth. Mutually exclusive " +
					"with username/password. May also be set via the CONFLUENCE_TOKEN environment variable.",
				Validators: []validator.String{
					stringvalidator.ConflictsWith(
						path.MatchRoot("username"),
						path.MatchRoot("password"),
					),
				},
			},
			"username": schema.StringAttribute{
				Optional: true,
				Description: "Username used together with password for HTTP Basic authentication. Mutually " +
					"exclusive with token. May also be set via the CONFLUENCE_USERNAME environment variable.",
				Validators: []validator.String{
					stringvalidator.AlsoRequires(path.MatchRoot("password")),
				},
			},
			"password": schema.StringAttribute{
				Optional:  true,
				Sensitive: true,
				Description: "Password used together with username for HTTP Basic authentication. May also be " +
					"set via the CONFLUENCE_PASSWORD environment variable.",
				Validators: []validator.String{
					stringvalidator.AlsoRequires(path.MatchRoot("username")),
				},
			},
			"skip_tls_verify": schema.BoolAttribute{
				Optional: true,
				Description: "Skip TLS certificate verification when connecting to the Confluence host. Only " +
					"use this for instances behind a trusted network with a self-signed certificate. Defaults to false.",
			},
		},
	}
}

func (p *confluenceDCProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config confluenceDCProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	host := firstNonEmpty(config.Host.ValueString(), os.Getenv("CONFLUENCE_HOST"))
	token := firstNonEmpty(config.Token.ValueString(), os.Getenv("CONFLUENCE_TOKEN"))
	username := firstNonEmpty(config.Username.ValueString(), os.Getenv("CONFLUENCE_USERNAME"))
	password := firstNonEmpty(config.Password.ValueString(), os.Getenv("CONFLUENCE_PASSWORD"))

	if host == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("host"),
			"Missing Confluence Host",
			"The provider requires a host, set either via the host attribute or the CONFLUENCE_HOST environment variable.",
		)
	}

	hasToken := token != ""
	hasBasicAuth := username != "" && password != ""

	switch {
	case hasToken && hasBasicAuth:
		resp.Diagnostics.AddError(
			"Conflicting Confluence Credentials",
			"Both a personal access token and username/password were configured (possibly from a mix of "+
				"provider configuration and environment variables). Configure only one authentication method.",
		)
	case !hasToken && !hasBasicAuth:
		resp.Diagnostics.AddError(
			"Missing Confluence Credentials",
			"The provider requires either a personal access token (token / CONFLUENCE_TOKEN) or a username "+
				"and password (username+password / CONFLUENCE_USERNAME+CONFLUENCE_PASSWORD).",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	c, err := client.New(client.Config{
		Host:          host,
		Token:         token,
		Username:      username,
		Password:      password,
		SkipTLSVerify: config.SkipTLSVerify.ValueBool(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Unable to Create Confluence Client", err.Error())
		return
	}

	resp.ResourceData = c
	resp.DataSourceData = c
}

func (p *confluenceDCProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewGroupResource,
		NewSpacePermissionResource,
	}
}

func (p *confluenceDCProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewSpaceDataSource,
	}
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

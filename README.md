# Terraform Provider for Confluence Data Center

A [Terraform](https://www.terraform.io) provider for managing [Confluence Data
Center](https://www.atlassian.com/software/confluence) (self-managed, on-prem)
instances. It is **not** compatible with Confluence Cloud, which has a
different REST API.

`confluencedc_group` and `confluencedc_space_permission` create/delete groups
and grant/revoke space permissions through REST endpoints under
`/rest/api/admin/group` and
`/rest/api/space/{spaceKey}/permissions/group/{groupName}` — separate from,
and unrelated to, the read-only `/rest/api/group` and
`/rest/api/space/{spaceKey}/permissions` endpoints, which is why those write
paths can look unsupported at first (they 404/405). If those write endpoints
aren't found (e.g. a Confluence Data Center release old enough to predate
them), the provider falls back to Confluence's legacy JSON-RPC API
(`confluenceservice-v2`), deprecated by Atlassian since Confluence 5.5 but
still present and functional on current releases.

## Requirements

- [Go](https://go.dev/doc/install) 1.27.1 (see `go.mod`)
- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.7
- A Confluence Data Center instance; `confluencedc_group` and `confluencedc_space_permission` use REST by default and only need the legacy JSON-RPC API enabled (the default) as a fallback on instances old enough to lack the REST write endpoints they rely on

## Troubleshooting

If `confluencedc_group` or `confluencedc_space_permission` fail with a
404/405 error mentioning the JSON-RPC endpoint, it means REST *and* the
JSON-RPC fallback both failed — most likely because an admin has disabled
the Remote API, which can happen as a side effect of a Confluence upgrade or
a general re-review of admin settings. Check **Confluence Administration >
General Configuration > Further Configuration** and make sure **"Remote API
(XML-RPC & SOAP)"** is checked; despite its name, this setting also gates
the JSON-RPC endpoint used as a fallback.

## Using the provider

```hcl
terraform {
  required_providers {
    confluencedc = {
      source  = "gustavebbers/confluence-dc"
      version = "~> 1.0"
    }
  }
}

provider "confluencedc" {}
```

See [`examples/provider`](./examples/provider) for a full provider
configuration example, and [`examples/resources`](./examples/resources) and
[`examples/data-sources`](./examples/data-sources) for examples of each
resource and data source.

## Authentication

The provider supports two mutually exclusive authentication methods:

- **Personal Access Token (PAT)** — recommended, especially for service
  accounts and CI/CD. Set the `token` provider attribute, or the
  `CONFLUENCE_TOKEN` environment variable.
- **HTTP Basic authentication** — set the `username` and `password` provider
  attributes together, or the `CONFLUENCE_USERNAME` and `CONFLUENCE_PASSWORD`
  environment variables.

The Confluence host is configured via the `host` attribute or the
`CONFLUENCE_HOST` environment variable. Configuring both a token and
username/password (from any combination of attributes and environment
variables) is an error.

## Resources and Data Sources

- `confluencedc_space` (data source) — reads an existing Confluence space by its key.
- `confluencedc_group` (resource) — manages a Confluence group.
- `confluencedc_space_permission` (resource) — grants a group a permission on a space.

> **Note:** `confluencedc_group` and `confluencedc_space_permission` grant/revoke
> through Confluence's legacy JSON-RPC API, since the REST API has no working
> write endpoints for these on Data Center. See the note at the top of this
> README.

## Developing the Provider

See [CONTRIBUTING.md](./CONTRIBUTING.md) for instructions on building the
provider locally, running tests, and generating documentation.

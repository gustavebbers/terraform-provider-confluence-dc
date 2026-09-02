# Terraform Provider for Confluence Data Center

A [Terraform](https://www.terraform.io) provider for managing [Confluence Data
Center](https://www.atlassian.com/software/confluence) (self-managed, on-prem)
instances. It is **not** compatible with Confluence Cloud, which has a
different REST API.

Confluence Data Center's REST API only supports *reading* groups and space
permissions, not creating or modifying them (verified empirically: it returns
404/405 for those write endpoints even on current releases). `confluencedc_group`
and `confluencedc_space_permission` therefore perform writes through
Confluence's legacy JSON-RPC API (`confluenceservice-v2`) instead. That API is
deprecated by Atlassian but still present and functional as of Confluence
Data Center 9.2; it must remain enabled on the target instance for these two
resources to work.

## Requirements

- [Go](https://go.dev/doc/install) 1.27.1 (see `go.mod`)
- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.7
- A Confluence Data Center instance with the legacy JSON-RPC API enabled (the default) for `confluencedc_group` and `confluencedc_space_permission`

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

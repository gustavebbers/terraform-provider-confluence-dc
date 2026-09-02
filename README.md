# Terraform Provider for Confluence Data Center

A [Terraform](https://www.terraform.io) provider for managing [Confluence Data
Center](https://www.atlassian.com/software/confluence) (self-managed, on-prem)
instances. It targets **Confluence Data Center 9.1+** and is **not** compatible
with Confluence Cloud, which has a different REST API.

## Requirements

- [Go](https://go.dev/doc/install) 1.27.1 (see `go.mod`)
- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.7
- Confluence Data Center 9.1 or later (see the note on `confluencedc_space_permission` below)

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

> **Note:** `confluencedc_space_permission` requires Confluence Data Center
> 9.1 or later, the first release to expose space permission management
> through the REST API. Earlier versions of Confluence Data Center have no
> supported way to manage space permissions via this provider.

## Developing the Provider

See [CONTRIBUTING.md](./CONTRIBUTING.md) for instructions on building the
provider locally, running tests, and generating documentation.

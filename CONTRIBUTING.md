# Contributing

Thank you for your interest in contributing to the Confluence Data Center
provider!

## Prerequisites

- [Go](https://go.dev/doc/install) (see `go.mod` for the minimum version)
- [Terraform](https://developer.hashicorp.com/terraform/install) >= 1.7
- [golangci-lint](https://golangci-lint.run/usage/install/) (for `make lint`)

## Building the provider

```shell
make build
```

This runs `go install .`, which compiles the provider and installs the
binary into `$GOPATH/bin`.

## Running tests

### Unit tests

Unit tests do not require any external services and can be run with:

```shell
make test
```

### Acceptance tests

Acceptance tests exercise the provider against a real Confluence Data
Center instance and will create/modify/destroy real resources. **They are
run separately from unit tests** because they require live infrastructure.

Requirements:

- A running Confluence Data Center instance with the legacy JSON-RPC API enabled (the default)
- `CONFLUENCE_HOST` set to the base URL of that instance
- Either:
  - `CONFLUENCE_TOKEN` (a personal access token), or
  - `CONFLUENCE_USERNAME` and `CONFLUENCE_PASSWORD`

```shell
export CONFLUENCE_HOST=https://confluence.example.com
export CONFLUENCE_TOKEN=xxxxx

make testacc
```

`make testacc` sets `TF_ACC=1` and runs `go test` with a 120-minute
timeout, since acceptance tests can be slow.

## Linting and formatting

```shell
make lint   # golangci-lint run
make fmt    # gofmt -s -w .
```

## Generating documentation

Provider documentation under `docs/` is generated from the schema
descriptions in `internal/provider` and the examples in `examples/` using
[tfplugindocs](https://github.com/hashicorp/terraform-plugin-docs). After
changing any schema description or example, regenerate the docs and
commit the result:

```shell
make generate
```

CI verifies that generated docs are up to date and will fail the build
if `go generate ./...` produces a diff that wasn't committed.

## Local development overrides

To test a locally-built provider against real Terraform configurations
without publishing it to a registry, use Terraform's
[`dev_overrides`](https://developer.terraform.io/terraform/cli/config/config-file#development-overrides-for-provider-developers)
feature. Development overrides bypass the usual provider installation
(and signature verification) entirely, so they work with unsigned local
builds.

1. Build and install the provider:

   ```shell
   make build
   ```

2. Find your `GOPATH`'s bin directory:

   ```shell
   go env GOPATH
   ```

3. Create (or edit) `~/.terraformrc` with a `dev_overrides` block pointing
   at that directory:

   ```hcl
   provider_installation {
     dev_overrides {
       "gustavebbers/confluence-dc" = "/Users/you/go/bin"
     }

     # For all other providers, install them as normal.
     direct {}
   }
   ```

   Replace `/Users/you/go/bin` with the actual output of `go env GOPATH`
   plus `/bin`.

4. Run `terraform plan`/`apply` as usual in a configuration that requires
   `gustavebbers/confluence-dc`. Terraform will print a warning that
   development overrides are in effect and use your local binary instead
   of downloading a release from the registry. You do **not** need to run
   `terraform init` while overrides are active (and `init` will actually
   fail complaining about the override, which is expected).

Remove or comment out the `dev_overrides` block when you're done testing
locally so Terraform goes back to using the published, signed provider
release.

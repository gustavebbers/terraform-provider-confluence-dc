terraform {
  required_providers {
    confluencedc = {
      source  = "gustavebbers/confluence-dc"
      version = "~> 1.0"
    }
  }
}

# Configuration is left empty here on purpose: the provider reads its
# settings from environment variables, which keeps credentials out of
# your Terraform configuration and state.
#
#   CONFLUENCE_HOST     - Base URL of the Confluence Data Center instance,
#                          e.g. https://confluence.example.com
#   CONFLUENCE_TOKEN     - Personal Access Token (PAT), recommended for
#                          service accounts and CI/CD.
#   CONFLUENCE_USERNAME  - Username for HTTP Basic auth (alternative to a
#                          token; requires CONFLUENCE_PASSWORD too).
#   CONFLUENCE_PASSWORD  - Password for HTTP Basic auth.
#
# token is mutually exclusive with username/password.
provider "confluencedc" {}

# Equivalent explicit configuration using a Personal Access Token (PAT).
# PATs are the recommended authentication method for Confluence Data
# Center 7.9+.
#
# provider "confluencedc" {
#   host  = "https://confluence.example.com"
#   token = var.confluence_token
# }

# Equivalent explicit configuration using HTTP Basic authentication.
# provider "confluencedc" {
#   host     = "https://confluence.example.com"
#   username = var.confluence_username
#   password = var.confluence_password
# }

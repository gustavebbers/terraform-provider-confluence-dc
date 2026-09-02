package provider

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// testAccProtoV6ProviderFactories are used to instantiate the provider
// during acceptance testing. The factory function is called for every
// Terraform CLI command executed to create a provider server to which the
// CLI can reattach.
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"confluencedc": providerserver.NewProtocol6WithError(New("test")()),
}

// testAccPreCheck validates that the required environment variables are set
// for acceptance tests to run against a real Confluence Data Center
// instance. It is called from the PreCheck field of every acceptance test's
// resource.TestCase.
func testAccPreCheck(t *testing.T) {
	t.Helper()

	if os.Getenv("CONFLUENCE_HOST") == "" {
		t.Fatal("CONFLUENCE_HOST must be set for acceptance tests")
	}

	hasToken := os.Getenv("CONFLUENCE_TOKEN") != ""
	hasBasicAuth := os.Getenv("CONFLUENCE_USERNAME") != "" && os.Getenv("CONFLUENCE_PASSWORD") != ""

	if !hasToken && !hasBasicAuth {
		t.Fatal("either CONFLUENCE_TOKEN, or both CONFLUENCE_USERNAME and CONFLUENCE_PASSWORD, must be set for acceptance tests")
	}
}

func TestFirstNonEmpty(t *testing.T) {
	tests := []struct {
		name   string
		values []string
		want   string
	}{
		{
			name:   "first value wins",
			values: []string{"a", "b"},
			want:   "a",
		},
		{
			name:   "skips leading empties",
			values: []string{"", "", "c"},
			want:   "c",
		},
		{
			name:   "all empty returns empty",
			values: []string{"", ""},
			want:   "",
		},
		{
			name:   "no values returns empty",
			values: nil,
			want:   "",
		},
		{
			name:   "single non-empty value",
			values: []string{"only"},
			want:   "only",
		},
		{
			name:   "empty string amid values still finds first non-empty",
			values: []string{"", "middle", "last"},
			want:   "middle",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := firstNonEmpty(tt.values...); got != tt.want {
				t.Errorf("firstNonEmpty(%v) = %q, want %q", tt.values, got, tt.want)
			}
		})
	}
}

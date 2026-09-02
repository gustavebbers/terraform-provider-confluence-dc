package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccGroupResource exercises the confluencedc_group resource's full
// lifecycle (create, read, import) against a real Confluence Data Center
// instance. It is gated by TF_ACC=1 (the standard terraform-plugin-testing
// convention) and by testAccPreCheck, which requires CONFLUENCE_HOST plus
// either CONFLUENCE_TOKEN or CONFLUENCE_USERNAME/CONFLUENCE_PASSWORD to be
// set. It does not run in CI or in a sandbox without a live Confluence
// instance; that is expected.
func TestAccGroupResource(t *testing.T) {
	groupName := "tf-acc-test-group"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing.
			{
				Config: testAccGroupResourceConfig(groupName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("confluencedc_group.test", "name", groupName),
					resource.TestCheckResourceAttr("confluencedc_group.test", "id", groupName),
				),
			},
			// ImportState testing.
			{
				ResourceName:      "confluencedc_group.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccGroupResourceConfig(name string) string {
	return fmt.Sprintf(`
resource "confluencedc_group" "test" {
  name = %[1]q
}
`, name)
}

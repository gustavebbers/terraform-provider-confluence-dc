package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccSpacePermissionResource exercises the confluencedc_space_permission
// resource's full lifecycle (create, read, import) against a real
// Confluence Data Center instance. Writes go through the legacy JSON-RPC
// API (see internal/client/jsonrpc.go), which must be enabled on the
// target instance.
//
// Granting a real permission requires a real, pre-existing space to target,
// so this test additionally requires CONFLUENCE_ACC_SPACE_KEY to be set to
// the key of a space the configured credentials are allowed to administer;
// it is skipped (not failed) when that variable is unset, since it is not
// something testAccPreCheck alone can validate.
func TestAccSpacePermissionResource(t *testing.T) {
	spaceKey := os.Getenv("CONFLUENCE_ACC_SPACE_KEY")
	if spaceKey == "" {
		t.Skip("CONFLUENCE_ACC_SPACE_KEY must be set to run TestAccSpacePermissionResource")
	}

	groupName := "tf-acc-test-space-permission-group"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing.
			{
				Config: testAccSpacePermissionResourceConfig(groupName, spaceKey, "read", "space"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("confluencedc_space_permission.test", "space_key", spaceKey),
					resource.TestCheckResourceAttr("confluencedc_space_permission.test", "group_name", groupName),
					resource.TestCheckResourceAttr("confluencedc_space_permission.test", "operation_key", "read"),
					resource.TestCheckResourceAttr("confluencedc_space_permission.test", "operation_target", "space"),
					resource.TestCheckResourceAttrSet("confluencedc_space_permission.test", "id"),
				),
			},
			// ImportState testing.
			{
				ResourceName:      "confluencedc_space_permission.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccSpacePermissionResourceConfig(groupName, spaceKey, operationKey, operationTarget string) string {
	return fmt.Sprintf(`
resource "confluencedc_group" "test" {
  name = %[1]q
}

resource "confluencedc_space_permission" "test" {
  space_key        = %[2]q
  group_name       = confluencedc_group.test.name
  operation_key    = %[3]q
  operation_target = %[4]q
}
`, groupName, spaceKey, operationKey, operationTarget)
}

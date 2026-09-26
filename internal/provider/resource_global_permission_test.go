package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccGlobalPermissionResource exercises the confluencedc_global_permission
// resource's full lifecycle (create, read, import) against a real Confluence
// Data Center instance. Unlike confluencedc_space_permission, there is no
// legacy JSON-RPC fallback for this resource, so it requires the REST
// endpoint (PUT /rest/api/permissions/group/{groupName}/grant) to be present
// on the target instance.
func TestAccGlobalPermissionResource(t *testing.T) {
	groupName := "tf-acc-test-global-permission-group"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing.
			{
				Config: testAccGlobalPermissionResourceConfig(groupName, "use", "application"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("confluencedc_global_permission.test", "group_name", groupName),
					resource.TestCheckResourceAttr("confluencedc_global_permission.test", "operation_key", "use"),
					resource.TestCheckResourceAttr("confluencedc_global_permission.test", "operation_target", "application"),
					resource.TestCheckResourceAttrSet("confluencedc_global_permission.test", "id"),
				),
			},
			// ImportState testing.
			{
				ResourceName:      "confluencedc_global_permission.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccGlobalPermissionResourceConfig(groupName, operationKey, operationTarget string) string {
	return fmt.Sprintf(`
resource "confluencedc_group" "test" {
  name = %[1]q
}

resource "confluencedc_global_permission" "test" {
  group_name       = confluencedc_group.test.name
  operation_key    = %[2]q
  operation_target = %[3]q
}
`, groupName, operationKey, operationTarget)
}

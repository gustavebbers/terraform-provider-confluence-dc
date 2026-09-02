package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccSpaceDataSource exercises the confluencedc_space data source
// against a real Confluence Data Center instance. It reads the space key to
// look up from CONFLUENCE_ACC_SPACE_KEY and is skipped (not failed) when
// that variable is unset, since it needs a real, pre-existing space to
// target.
func TestAccSpaceDataSource(t *testing.T) {
	spaceKey := os.Getenv("CONFLUENCE_ACC_SPACE_KEY")
	if spaceKey == "" {
		t.Skip("CONFLUENCE_ACC_SPACE_KEY must be set to run TestAccSpaceDataSource")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSpaceDataSourceConfig(spaceKey),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.confluencedc_space.test", "key", spaceKey),
					resource.TestCheckResourceAttrSet("data.confluencedc_space.test", "id"),
					resource.TestCheckResourceAttrSet("data.confluencedc_space.test", "name"),
				),
			},
		},
	})
}

func testAccSpaceDataSourceConfig(key string) string {
	return fmt.Sprintf(`
data "confluencedc_space" "test" {
  key = %[1]q
}
`, key)
}

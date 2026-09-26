resource "confluencedc_group" "confluence_license" {
  name = "confluence-license"
}

# Give everyone in this group standard, instance-wide access to log in and
# use Confluence.
resource "confluencedc_global_permission" "confluence_license_use" {
  group_name       = confluencedc_group.confluence_license.name
  operation_key    = "use"
  operation_target = "application"
}

resource "confluencedc_group" "platform_admin" {
  name = "platform-admin"
}

# Give this group full system administrator rights.
resource "confluencedc_global_permission" "platform_admin_system_administer" {
  group_name       = confluencedc_group.platform_admin.name
  operation_key    = "administer"
  operation_target = "system"
}

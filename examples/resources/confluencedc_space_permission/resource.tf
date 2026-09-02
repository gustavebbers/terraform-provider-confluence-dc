data "confluencedc_space" "engineering" {
  key = "ENG"
}

resource "confluencedc_group" "developers" {
  name = "developers"
}

# Allow the "developers" group to view the space.
resource "confluencedc_space_permission" "developers_read" {
  space_key        = data.confluencedc_space.engineering.key
  group_name       = confluencedc_group.developers.name
  operation_key    = "read"
  operation_target = "space"
}

# Allow the "developers" group to create new pages in the space.
resource "confluencedc_space_permission" "developers_create_page" {
  space_key        = data.confluencedc_space.engineering.key
  group_name       = confluencedc_group.developers.name
  operation_key    = "create"
  operation_target = "page"
}

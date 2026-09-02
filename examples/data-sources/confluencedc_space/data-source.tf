data "confluencedc_space" "engineering" {
  key = "ENG"
}

output "engineering_space_name" {
  value = data.confluencedc_space.engineering.name
}

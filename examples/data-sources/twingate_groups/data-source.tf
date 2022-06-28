data "twingate_groups" "foo" {
  name = "<your group's name>"
}

# Group names are not constrained to be unique within Twingate,
# so it is possible that this data source will return multiple list items.
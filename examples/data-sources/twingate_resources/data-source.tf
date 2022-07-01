data "twingate_resources" "foo" {
  name = "<your resource's name>"
}

# Resource names are not constrained to be unique within Twingate,
# so it is possible that this data source will return multiple list items.
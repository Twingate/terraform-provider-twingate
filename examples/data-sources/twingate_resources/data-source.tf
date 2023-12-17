data "twingate_resources" "foo" {
  name = "<your resource's name>"
#  name_regexp = "<your resource's name>"
#  name_contains = "<your resource's name>"
#  name_exclude = "<your resource's name>"
#  name_prefix = "<your resource's name>"
#  name_suffix = "<your resource's name>"
}

# Resource names are not constrained to be unique within Twingate,
# so it is possible that this data source will return multiple list items.
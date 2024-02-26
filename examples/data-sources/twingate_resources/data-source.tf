data "twingate_resources" "foo" {
  name = "<your resource's name>"
  #  name_regexp = "<regular expression of resource name>"
  #  name_contains = "<a string in the resource name>"
  #  name_exclude_contains = "<your resource's name to exclude>"
  #  name_prefix = "<prefix of resource name>"
  #  name_suffix = "<suffix of resource name>"
}

# Resource names are not constrained to be unique within Twingate,
# so it is possible that this data source will return multiple list items.
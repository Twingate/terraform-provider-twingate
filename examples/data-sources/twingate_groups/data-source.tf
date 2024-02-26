data "twingate_groups" "foo" {
  name = "<your group's name>"
  #  name_regexp = "<regular expression of group name>"
  #  name_contains = "<a string in the group name>"
  #  name_exclude = "<your group's name to exclude>"
  #  name_prefix = "<prefix of resource name>"
  #  name_suffix = "<suffix of resource name>"
}

# Group names are not constrained to be unique within Twingate,
# so it is possible that this data source will return multiple list items.
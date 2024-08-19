provider "twingate" {
  api_token = "1234567890abcdef"
  network   = "mynetwork"
}

resource "twingate_group" "example1" {
  name = "example_1"
}

resource "twingate_group" "example2" {
  name = "example_2"
}

data "twingate_groups" "example" {
  name_prefix = "example"

  depends_on = [twingate_group.example1, twingate_group.example2]
}

resource "twingate_dns_filtering_profile" "example" {
  name = "Example DNS Filtering Profile"
  priority = 2
  fallback_method = "AUTO"
  groups = toset(data.twingate_groups.example.groups[*].id)

  allowed_domains {
    is_authoritative = false
    domains = [
      "twingate.com",
      "zoom.us"
    ]
  }

  denied_domains {
    is_authoritative = true
    domains = [
      "evil.example"
    ]
  }

  content_categories {
    block_adult_content = true
  }

  security_categories {
    block_dns_rebinding = false
    block_newly_registered_domains = false
  }

  privacy_categories {
    block_disguised_trackers = true
  }

}


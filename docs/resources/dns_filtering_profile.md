---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "twingate_dns_filtering_profile Resource - terraform-provider-twingate"
subcategory: ""
description: |-
  DNS filtering gives you the ability to control what websites your users can access. DNS filtering is only available on certain plans. For more information, see Twingate's documentation https://www.twingate.com/docs/dns-filtering. DNS filtering must be enabled for this resources to work. If DNS filtering isn't enabled, the provider will throw an error.
---

# twingate_dns_filtering_profile (Resource)

DNS filtering gives you the ability to control what websites your users can access. DNS filtering is only available on certain plans. For more information, see Twingate's [documentation](https://www.twingate.com/docs/dns-filtering). DNS filtering must be enabled for this resources to work. If DNS filtering isn't enabled, the provider will throw an error.

## Example Usage

```terraform
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
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `name` (String) The DNS filtering profile's name.
- `priority` (Number) A floating point number representing the profile's priority.

### Optional

- `allowed_domains` (Block, Optional) A block with the following attributes. (see [below for nested schema](#nestedblock--allowed_domains))
- `content_categories` (Block, Optional) A block with the following attributes. (see [below for nested schema](#nestedblock--content_categories))
- `denied_domains` (Block, Optional) A block with the following attributes. (see [below for nested schema](#nestedblock--denied_domains))
- `fallback_method` (String) The DNS filtering profile's fallback method. One of "AUTO" or "STRICT". Defaults to "STRICT".
- `groups` (Set of String) A set of group IDs that have this as their DNS filtering profile. Defaults to an empty set.
- `privacy_categories` (Block, Optional) A block with the following attributes. (see [below for nested schema](#nestedblock--privacy_categories))
- `security_categories` (Block, Optional) A block with the following attributes. (see [below for nested schema](#nestedblock--security_categories))

### Read-Only

- `id` (String) Autogenerated ID of the DNS filtering profile.

<a id="nestedblock--allowed_domains"></a>
### Nested Schema for `allowed_domains`

Optional:

- `domains` (Set of String) A set of allowed domains. Defaults to an empty set.
- `is_authoritative` (Boolean) Whether Terraform should override changes made outside of Terraform. Defaults to true.


<a id="nestedblock--content_categories"></a>
### Nested Schema for `content_categories`

Optional:

- `block_adult_content` (Boolean) Whether to block adult content. Defaults to false.
- `block_dating` (Boolean) Whether to block dating content. Defaults to false.
- `block_gambling` (Boolean) Whether to block gambling content. Defaults to false.
- `block_games` (Boolean) Whether to block games. Defaults to false.
- `block_piracy` (Boolean) Whether to block piracy sites. Defaults to false.
- `block_social_media` (Boolean) Whether to block social media. Defaults to false.
- `block_streaming` (Boolean) Whether to block streaming content. Defaults to false.
- `enable_safesearch` (Boolean) Whether to force safe search. Defaults to false.
- `enable_youtube_restricted_mode` (Boolean) Whether to force YouTube to use restricted mode. Defaults to false.


<a id="nestedblock--denied_domains"></a>
### Nested Schema for `denied_domains`

Optional:

- `domains` (Set of String) A set of denied domains. Defaults to an empty set.
- `is_authoritative` (Boolean) Whether Terraform should override changes made outside of Terraform. Defaults to true.


<a id="nestedblock--privacy_categories"></a>
### Nested Schema for `privacy_categories`

Optional:

- `block_ads_and_trackers` (Boolean) Whether to block ads and trackers. Defaults to false.
- `block_affiliate_links` (Boolean) Whether to block affiliate links. Defaults to false.
- `block_disguised_trackers` (Boolean) Whether to block disguised third party trackers. Defaults to false.


<a id="nestedblock--security_categories"></a>
### Nested Schema for `security_categories`

Optional:

- `block_cryptojacking` (Boolean) Whether to block cryptojacking sites. Defaults to true.
- `block_dns_rebinding` (Boolean) Blocks public DNS entries from returning private IP addresses. Defaults to true.
- `block_domain_generation_algorithms` (Boolean) Blocks DGA domains. Defaults to true.
- `block_idn_homoglyph` (Boolean) Whether to block homoglyph attacks. Defaults to true.
- `block_newly_registered_domains` (Boolean) Blocks newly registered domains. Defaults to true.
- `block_parked_domains` (Boolean) Block parked domains. Defaults to true.
- `block_typosquatting` (Boolean) Blocks typosquatted domains. Defaults to true.
- `enable_google_safe_browsing` (Boolean) Whether to use Google Safe browsing lists to block content. Defaults to true.
- `enable_threat_intelligence_feeds` (Boolean) Whether to filter content using threat intelligence feeds. Defaults to true.

## Import

Import is supported using the following syntax:

```shell
terraform import twingate_dns_filtering_profile.example RG5zRmlsdGVyaW5nUHJvZmlsZToxY2I4YzM0YTc0
```

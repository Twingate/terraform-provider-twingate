package query

type UpdateDNSFilteringProfile struct {
	DNSFilteringProfileEntityResponse `graphql:"dnsFilteringProfileUpdate(id: $id, name: $name, priority: $priority, allowedDomains: $allowedDomains, deniedDomains: $deniedDomains, fallbackMethod: $fallbackMethod, groups: $groups, privacyCategoryConfig: $privacyCategoryConfig, securityCategoryConfig: $securityCategoryConfig, contentCategoryConfig: $contentCategoryConfig)"`
}

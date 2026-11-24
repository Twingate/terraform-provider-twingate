package query

import "github.com/Twingate/terraform-provider-twingate/v5/twingate/internal/model"

//nolint:lll
type CreateWebAppResource struct {
	WebAppResourceEntityResponse `graphql:"webAppResourceCreate(name: $name, address: $address, gatewayId: $gatewayId, remoteNetworkId: $remoteNetworkId, isVisible: $isVisible, alias: $alias, securityPolicyId: $securityPolicyId, tags: $tags, upstream: $upstream, downstream: $downstream, requestHeaderRewrites: $requestHeaderRewrites, accessPolicy: $accessPolicy, approvalMode: $approvalMode)"`
}

func (q CreateWebAppResource) IsEmpty() bool {
	return q.Entity == nil
}

func (q CreateWebAppResource) ToModel() *model.WebAppResource {
	if q.Entity == nil {
		return nil
	}

	return q.Entity.ToModel()
}

type WebAppResourceEntityResponse struct {
	Entity *gqlWebAppResource
	OkError
}

type KeyValuePair struct {
	Key   string
	Value string
}

type WebAppUpstream struct {
	Port    int64
	TLSMode string
}

type WebAppDownstream struct {
	Port    int64
	TLSMode string
}

type gqlWebAppResource struct {
	gqlPAMResource

	Upstream              WebAppUpstream
	Downstream            WebAppDownstream
	RequestHeaderRewrites []KeyValuePair
}

func (g gqlWebAppResource) ToModel() *model.WebAppResource {
	return &model.WebAppResource{
		PAMResource:           g.gqlPAMResource.ToModel(),
		Upstream:              model.WebAppUpstream{Port: g.Upstream.Port, TLSMode: g.Upstream.TLSMode},
		Downstream:            model.WebAppDownstream{Port: g.Downstream.Port, TLSMode: g.Downstream.TLSMode},
		RequestHeaderRewrites: headerRewritesToModel(g.RequestHeaderRewrites),
	}
}

// headerRewritesToModel returns nil for an empty list so that a resource without
// header rewrites is indistinguishable from one where the API dropped the field.
func headerRewritesToModel(pairs []KeyValuePair) map[string]string {
	if len(pairs) == 0 {
		return nil
	}

	rewrites := make(map[string]string, len(pairs))
	for _, pair := range pairs {
		rewrites[pair.Key] = pair.Value
	}

	return rewrites
}

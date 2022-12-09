package query

import (
	"fmt"
	"time"

	"github.com/Twingate/terraform-provider-twingate/twingate/internal/model"
	"github.com/twingate/go-graphql-client"
)

type ReadServiceAccountKey struct {
	ServiceAccountKey *gqlServiceKey `graphql:"serviceAccountKey(id: $id)"`
}

func (q ReadServiceAccountKey) ToModel() (*model.ServiceKey, error) {
	if q.ServiceAccountKey == nil {
		return nil, nil //nolint
	}

	return q.ServiceAccountKey.ToModel()
}

type gqlServiceKey struct {
	IDName
	ExpiresAt      graphql.String
	Status         graphql.String
	ServiceAccount gqlServiceAccount
}

func (q gqlServiceKey) ToModel() (*model.ServiceKey, error) {
	expirationTime, err := q.parseExpirationTime()
	if err != nil {
		return nil, err
	}

	return &model.ServiceKey{
		ID:             q.StringID(),
		Name:           q.StringName(),
		Service:        q.ServiceAccount.StringID(),
		ExpirationTime: expirationTime,
		Status:         string(q.Status),
	}, nil
}

func (q gqlServiceKey) parseExpirationTime() (int, error) {
	if q.ExpiresAt == "" {
		return 0, nil
	}

	expiresAt, err := time.Parse(time.RFC3339, string(q.ExpiresAt))
	if err != nil {
		return -1, fmt.Errorf("failed to parse expiration time `%s`: %w", q.ExpiresAt, err)
	}

	return getDaysTillExpiration(expiresAt), nil
}

func getDaysTillExpiration(expiresAt time.Time) int {
	const hoursInDay = 24

	return int(time.Until(expiresAt).Hours()/hoursInDay) + 1
}

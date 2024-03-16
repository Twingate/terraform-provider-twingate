package query

import (
	"fmt"
	"time"

	"github.com/Twingate/terraform-provider-twingate/v2/twingate/internal/model"
)

type ReadServiceAccountKey struct {
	ServiceAccountKey *gqlServiceKey `graphql:"serviceAccountKey(id: $id)"`
}

func (q ReadServiceAccountKey) IsEmpty() bool {
	return q.ServiceAccountKey == nil
}

func (q ReadServiceAccountKey) ToModel() (*model.ServiceKey, error) {
	if q.ServiceAccountKey == nil {
		return nil, nil //nolint
	}

	return q.ServiceAccountKey.ToModel()
}

type gqlServiceKey struct {
	IDName
	ExpiresAt      string
	Status         string
	ServiceAccount gqlServiceAccount
}

func (q gqlServiceKey) ToModel() (*model.ServiceKey, error) {
	expirationTime, err := q.parseExpirationTime()
	if err != nil {
		return nil, err
	}

	return &model.ServiceKey{
		ID:             string(q.ID),
		Name:           q.Name,
		Service:        string(q.ServiceAccount.ID),
		ExpirationTime: expirationTime,
		Status:         q.Status,
	}, nil
}

func (q gqlServiceKey) parseExpirationTime() (int, error) {
	if q.ExpiresAt == "" {
		return 0, nil
	}

	expiresAt, err := time.Parse(time.RFC3339, q.ExpiresAt)
	if err != nil {
		return -1, fmt.Errorf("failed to parse expiration time `%s`: %w", q.ExpiresAt, err)
	}

	return getDaysTillExpiration(expiresAt), nil
}

func getDaysTillExpiration(expiresAt time.Time) int {
	const hoursInDay = 24

	return int(time.Until(expiresAt).Hours()/hoursInDay) + 1
}

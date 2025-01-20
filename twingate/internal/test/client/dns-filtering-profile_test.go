package client

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/model"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
)

func TestClientDNSProfileCreateOk(t *testing.T) {
	t.Run("Test Twingate Resource : Create DNS Profile Ok", func(t *testing.T) {
		expected := &model.DNSFilteringProfile{
			ID:             "test-id",
			Name:           "test",
			Priority:       2,
			Groups:         []string{},
			FallbackMethod: "STRICT",
		}

		jsonResponse := `{
		  "data": {
		    "dnsFilteringProfileCreate": {
		      "entity": {
		        "id": "test-id",
		        "name": "test",
		        "priority": 2,
		        "fallbackMethod": "STRICT"
		      },
		      "ok": true,
		      "error": null
		    }
		  }
		}`

		c := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", c.GraphqlServerURL,
			httpmock.NewStringResponder(200, jsonResponse))

		profile, err := c.CreateDNSFilteringProfile(context.Background(), "test")

		assert.NoError(t, err)
		assert.EqualValues(t, expected, profile)
	})
}

func TestClientDNSProfileCreateError(t *testing.T) {
	t.Run("Test Twingate Resource : Create DNS Profile Error", func(t *testing.T) {
		jsonResponse := `{
		  "data": {
		    "dnsFilteringProfileCreate": {
		      "ok": false,
		      "error": "error_1"
		    }
		  }
		}`

		c := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", c.GraphqlServerURL,
			httpmock.NewStringResponder(200, jsonResponse))

		profile, err := c.CreateDNSFilteringProfile(context.Background(), "test")

		assert.EqualError(t, err, "failed to create DNS filtering profile with name test: error_1")
		assert.Nil(t, profile)
	})
}

func TestClientDNSProfileCreateRequestError(t *testing.T) {
	t.Run("Test Twingate Resource : Create DNS Profile Request Error", func(t *testing.T) {
		c := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()

		httpmock.RegisterResponder("POST", c.GraphqlServerURL,
			httpmock.NewErrorResponder(errBadRequest))

		profile, err := c.CreateDNSFilteringProfile(context.Background(), "test")

		assert.EqualError(t, err, graphqlErr(c, "failed to create DNS filtering profile with name test", errBadRequest))
		assert.Nil(t, profile)
	})
}

func TestClientCreateEmptyDNSProfileError(t *testing.T) {
	t.Run("Test Twingate Resource : Create Empty DNS Profile Error", func(t *testing.T) {
		jsonResponse := `{
		  "data": {
		    "dnsFilteringProfileCreate": null
		  }
		}`

		c := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", c.GraphqlServerURL,
			httpmock.NewStringResponder(200, jsonResponse))

		profile, err := c.CreateDNSFilteringProfile(context.Background(), "")

		assert.EqualError(t, err, "failed to create DNS filtering profile: name is empty")
		assert.Nil(t, profile)
	})
}

func TestClientDNSProfileUpdateOk(t *testing.T) {
	t.Run("Test Twingate Resource : Update DNS Profile Ok", func(t *testing.T) {
		jsonResponse := `{
		  "data": {
		    "dnsFilteringProfileUpdate": {
		      "entity": {
		        "id": "id",
		        "name": "test"
		      },
		      "ok": true,
		      "error": null
		    }
		  }
		}`

		c := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", c.GraphqlServerURL,
			httpmock.NewStringResponder(200, jsonResponse))

		_, err := c.UpdateDNSFilteringProfile(context.Background(), &model.DNSFilteringProfile{
			ID:   "id",
			Name: "test",
			PrivacyCategories: &model.PrivacyCategories{
				BlockAffiliate:         true,
				BlockAdsAndTrackers:    true,
				BlockDisguisedTrackers: true,
			},
			SecurityCategories: &model.SecurityCategory{
				EnableThreatIntelligenceFeeds:   true,
				EnableGoogleSafeBrowsing:        true,
				BlockCryptojacking:              true,
				BlockIdnHomographs:              true,
				BlockTyposquatting:              true,
				BlockDNSRebinding:               true,
				BlockNewlyRegisteredDomains:     true,
				BlockDomainGenerationAlgorithms: true,
				BlockParkedDomains:              true,
			},
			ContentCategories: &model.ContentCategory{
				BlockGambling:               true,
				BlockDating:                 true,
				BlockAdultContent:           true,
				BlockSocialMedia:            true,
				BlockGames:                  true,
				BlockStreaming:              true,
				BlockPiracy:                 true,
				EnableYoutubeRestrictedMode: true,
				EnableSafeSearch:            true,
			},
		})

		assert.NoError(t, err)
	})
}

func TestClientDNSProfileUpdateError(t *testing.T) {
	t.Run("Test Twingate Resource : Update DNS Profile Error", func(t *testing.T) {
		jsonResponse := `{
		  "data": {
		    "dnsFilteringProfileUpdate": {
		      "ok": false,
		      "error": "error_1"
		    }
		  }
		}`

		c := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", c.GraphqlServerURL,
			httpmock.NewStringResponder(200, jsonResponse))

		const profileId = "g1"
		_, err := c.UpdateDNSFilteringProfile(context.Background(), &model.DNSFilteringProfile{ID: profileId, Name: "test"})

		assert.EqualError(t, err, fmt.Sprintf("failed to update DNS filtering profile with id %s: error_1", profileId))
	})
}

func TestClientDNSProfileUpdateRequestError(t *testing.T) {
	t.Run("Test Twingate Resource : Update DNS Profile Request Error", func(t *testing.T) {
		c := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", c.GraphqlServerURL,
			httpmock.NewErrorResponder(errBadRequest))

		const profileId = "g1"
		_, err := c.UpdateDNSFilteringProfile(context.Background(), &model.DNSFilteringProfile{ID: profileId, Name: "test"})

		assert.EqualError(t, err, graphqlErr(c, "failed to update DNS filtering profile with id "+profileId, errBadRequest))
	})
}

func TestClientDNSProfileUpdateEmptyResponse(t *testing.T) {
	t.Run("Test Twingate Resource : Update DNS Profile - Empty Response", func(t *testing.T) {
		jsonResponse := `{
		  "data": {
		    "dnsFilteringProfileUpdate": {
		      "ok": true,
		      "entity": null
		    }
		  }
		}`

		c := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", c.GraphqlServerURL,
			httpmock.NewStringResponder(200, jsonResponse))

		const profileId = "g1"
		_, err := c.UpdateDNSFilteringProfile(context.Background(), &model.DNSFilteringProfile{ID: profileId, Name: "test"})

		assert.EqualError(t, err, fmt.Sprintf("failed to update DNS filtering profile with id %s: query result is empty", profileId))
	})
}

func TestClientDNSProfileUpdateWithEmptyID(t *testing.T) {
	t.Run("Test Twingate Resource : Update DNS Profile With Empty ID", func(t *testing.T) {
		c := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()

		_, err := c.UpdateDNSFilteringProfile(context.Background(), &model.DNSFilteringProfile{Name: "test"})

		assert.EqualError(t, err, "failed to update DNS filtering profile: id is empty")
	})
}

func TestClientDNSProfileReadOk(t *testing.T) {
	t.Run("Test Twingate Resource : Read DNS Profile Ok", func(t *testing.T) {
		expected := &model.DNSFilteringProfile{
			ID:     "id",
			Name:   "name",
			Groups: []string{},
		}

		jsonResponse := `{
		  "data": {
		    "dnsFilteringProfile": {
		      "id": "id",
		      "name": "name"
		    }
		  }
		}`

		c := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", c.GraphqlServerURL,
			httpmock.NewStringResponder(200, jsonResponse))

		profile, err := c.ReadDNSFilteringProfile(context.Background(), "id")

		assert.NoError(t, err)
		assert.Equal(t, expected, profile)
	})
}

func TestClientDNSProfileReadError(t *testing.T) {
	t.Run("Test Twingate Resource : Read DNS Profile Error", func(t *testing.T) {
		jsonResponse := `{
		  "data": {
		    "dnsFilteringProfile": null
		  }
		}`

		c := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", c.GraphqlServerURL,
			httpmock.NewStringResponder(200, jsonResponse))

		const profileId = "g1"
		profile, err := c.ReadDNSFilteringProfile(context.Background(), profileId)

		assert.Nil(t, profile)
		assert.EqualError(t, err, fmt.Sprintf("failed to read DNS filtering profile with id %s: query result is empty", profileId))
	})
}

func TestClientDNSProfileReadRequestError(t *testing.T) {
	t.Run("Test Twingate Resource : Read DNS Profile Request Error", func(t *testing.T) {
		c := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", c.GraphqlServerURL,
			httpmock.NewErrorResponder(errBadRequest))

		const profileId = "g1"
		profile, err := c.ReadDNSFilteringProfile(context.Background(), profileId)

		assert.Nil(t, profile)
		assert.EqualError(t, err, graphqlErr(c, "failed to read DNS filtering profile with id "+profileId, errBadRequest))
	})
}

func TestClientReadEmptyDNSProfileError(t *testing.T) {
	t.Run("Test Twingate Resource : Read Empty DNS Profile Error", func(t *testing.T) {
		jsonResponse := `{
		  "data": {
		    "dnsFilteringProfile": null
		  }
		}`

		c := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", c.GraphqlServerURL,
			httpmock.NewStringResponder(200, jsonResponse))

		profile, err := c.ReadDNSFilteringProfile(context.Background(), "")

		assert.EqualError(t, err, "failed to read DNS filtering profile: id is empty")
		assert.Nil(t, profile)
	})
}

func TestClientDNSProfileReadErrorOnFetchPages(t *testing.T) {
	t.Run("Test Twingate Resource : Read DNS Profile Error On Fetch Pages", func(t *testing.T) {
		jsonResponse := `{
          "data": {
            "dnsFilteringProfile": {
                "id": "test-id",
                "name": "name",
                "groups": {
                  "pageInfo": {
                    "endCursor": "cursor-001",
                    "hasNextPage": true
                  },
                  "edges": [
                    {
                      "node": {
                        "id": "group-id"
                      }
                    }
                  ]
                }
            }
          }
        }`

		c := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", c.GraphqlServerURL,
			MultipleResponders(
				httpmock.NewStringResponder(http.StatusOK, jsonResponse),
				httpmock.NewErrorResponder(errBadRequest),
			))

		_, err := c.ReadDNSFilteringProfile(context.Background(), "test-id")

		assert.EqualError(t, err, graphqlErr(c, "failed to read group with id All", errBadRequest))
	})
}

func TestClientDNSProfileReadEmptyOnFetchPages(t *testing.T) {
	t.Run("Test Twingate Resource : Read DNS Profile Error On Fetch Pages", func(t *testing.T) {
		response1 := `{
          "data": {
            "dnsFilteringProfile": {
                "id": "test-id",
                "name": "name",
                "groups": {
                  "pageInfo": {
                    "endCursor": "cursor-001",
                    "hasNextPage": true
                  },
                  "edges": [
                    {
                      "node": {
                        "id": "profile-id"
                      }
                    }
                  ]
                }
            }
          }
        }`

		response2 := `{
          "data": {
            "dnsFilteringProfile": null
          }
        }`

		c := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", c.GraphqlServerURL,
			MultipleResponders(
				httpmock.NewStringResponder(http.StatusOK, response1),
				httpmock.NewStringResponder(http.StatusOK, response2),
			))

		profile, err := c.ReadDNSFilteringProfile(context.Background(), "test-id")

		assert.Nil(t, profile)
		assert.EqualError(t, err, fmt.Sprintf(`failed to read group with id All: query result is empty`))
	})
}

func TestClientDNSProfileReadOkOnFetchPages(t *testing.T) {
	t.Run("Test Twingate Resource : Read DNS Profile Error On Fetch Pages", func(t *testing.T) {
		expected := &model.DNSFilteringProfile{
			ID:     "profile-id",
			Name:   "name",
			Groups: []string{"group-1", "group-2"},
		}

		response1 := `{
          "data": {
            "dnsFilteringProfile": {
                "id": "profile-id",
                "name": "name",
                "groups": {
                  "pageInfo": {
                    "endCursor": "cursor-001",
                    "hasNextPage": true
                  },
                  "edges": [
                    {
                      "node": {
                        "id": "group-1"
                      }
                    }
                  ]
                }
            }
          }
        }`

		response2 := `{
          "data": {
            "dnsFilteringProfile": {
                "id": "profile-id",
                "name": "name",
                "groups": {
                  "pageInfo": {
                    "endCursor": "cursor-001",
                    "hasNextPage": false
                  },
                  "edges": [
                    {
                      "node": {
                        "id": "group-2"
                      }
                    }
                  ]
                }
            }
          }
        }`

		c := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", c.GraphqlServerURL,
			MultipleResponders(
				httpmock.NewStringResponder(http.StatusOK, response1),
				httpmock.NewStringResponder(http.StatusOK, response2),
			))

		profile, err := c.ReadDNSFilteringProfile(context.Background(), "profile-id")

		assert.NoError(t, err)
		assert.Equal(t, expected, profile)
	})
}

func TestClientDeleteDNSProfileOk(t *testing.T) {
	t.Run("Test Twingate Resource : Delete DNS Profile Ok", func(t *testing.T) {
		jsonResponse := `{
		  "data": {
		    "dnsFilteringProfileDelete": {
		      "ok": true,
		      "error": null
		    }
		  }
		}`

		c := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", c.GraphqlServerURL,
			httpmock.NewStringResponder(200, jsonResponse))

		err := c.DeleteDNSFilteringProfile(context.Background(), "profile-id")

		assert.NoError(t, err)
	})
}

func TestClientDeleteEmptyDNSProfileError(t *testing.T) {
	t.Run("Test Twingate Resource : Delete Empty DNS Profile Error", func(t *testing.T) {
		jsonResponse := `{
		  "data": {
		    "dnsFilteringProfileDelete": null
		  }
		}`

		c := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", c.GraphqlServerURL,
			httpmock.NewStringResponder(200, jsonResponse))

		err := c.DeleteDNSFilteringProfile(context.Background(), "")

		assert.EqualError(t, err, "failed to delete DNS filtering profile: id is empty")
	})
}

func TestClientDeleteDNSProfileError(t *testing.T) {
	t.Run("Test Twingate Resource : Delete DNS Profile Error", func(t *testing.T) {
		jsonResponse := `{
		  "data": {
		    "dnsFilteringProfileDelete": {
		      "ok": false,
		      "error": "error_1"
		    }
		  }
		}`

		c := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", c.GraphqlServerURL,
			httpmock.NewStringResponder(200, jsonResponse))

		err := c.DeleteDNSFilteringProfile(context.Background(), "profile-id")

		assert.EqualError(t, err, "failed to delete DNS filtering profile with id profile-id: error_1")
	})
}

func TestClientDeleteDNSProfileRequestError(t *testing.T) {
	t.Run("Test Twingate Resource : Delete DNS Profile Request Error", func(t *testing.T) {
		c := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", c.GraphqlServerURL,
			httpmock.NewErrorResponder(errBadRequest))

		err := c.DeleteDNSFilteringProfile(context.Background(), "profile-id")

		assert.EqualError(t, err, graphqlErr(c, "failed to delete DNS filtering profile with id profile-id", errBadRequest))
	})
}

func TestClientDNSProfilesReadOk(t *testing.T) {
	t.Run("Test Twingate Resource : Read DNS Profiles Ok", func(t *testing.T) {
		expected := []*model.DNSFilteringProfile{
			{
				ID:   "id1",
				Name: "profile1",
			},
			{
				ID:   "id2",
				Name: "profile2",
			},
			{
				ID:   "id3",
				Name: "profile3",
			},
		}

		jsonResponse := `{
		  "data": {
		    "dnsFilteringProfiles": [
		        {
		            "id": "id1",
		            "name": "profile1"
		        },
		        {
		            "id": "id2",
		            "name": "profile2"
		        },
		        {
		            "id": "id3",
		            "name": "profile3"
		        }
		      ]
		    }
		  }`

		c := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", c.GraphqlServerURL,
			httpmock.NewStringResponder(200, jsonResponse))

		profiles, err := c.ReadShallowDNSFilteringProfiles(context.Background())

		assert.NoError(t, err)
		assert.Equal(t, expected, profiles)
	})
}

func TestClientDNSProfilesReadError(t *testing.T) {
	t.Run("Test Twingate Resource : Read DNS Profile Error", func(t *testing.T) {
		emptyResponse := `{
		  "data": {
		    "dnsFilteringProfiles": null
		  }
		}`

		c := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", c.GraphqlServerURL,
			httpmock.NewStringResponder(200, emptyResponse))

		profiles, err := c.ReadShallowDNSFilteringProfiles(context.Background())

		assert.Nil(t, profiles)
		assert.EqualError(t, err, "failed to read DNS filtering profile with id All: query result is empty")
	})
}

func TestClientDNSProfilesReadRequestError(t *testing.T) {
	t.Run("Test Twingate Resource : Read DNS Profiles Request Error", func(t *testing.T) {
		c := newHTTPMockClient()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("POST", c.GraphqlServerURL,
			httpmock.NewErrorResponder(errBadRequest))

		profiles, err := c.ReadShallowDNSFilteringProfiles(context.Background())

		assert.Nil(t, profiles)
		assert.EqualError(t, err, graphqlErr(c, "failed to read DNS filtering profile with id All", errBadRequest))
	})
}

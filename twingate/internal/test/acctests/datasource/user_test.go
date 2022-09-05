package datasource

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"regexp"
	"testing"

	"github.com/Twingate/terraform-provider-twingate/twingate/internal/test/acctests"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/transport"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDatasourceTwingateUser_basic(t *testing.T) {
	t.Parallel()
	t.SkipNow() // fixed in other PR
	t.Run("Test Twingate Datasource : Acc User Basic", func(t *testing.T) {
		user, err := getTestUser()
		if err != nil {
			t.Skip("can't run test:", err)
		}

		resource.Test(t, resource.TestCase{
			ProviderFactories: acctests.ProviderFactories,
			PreCheck:          func() { acctests.PreCheck(t) },
			Steps: []resource.TestStep{
				{
					Config: testDatasourceTwingateUser(user.ID),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckOutput("my_user_email", user.Email),
					),
				},
			},
		})
	})
}

func getTestUser() (*transport.User, error) {
	if acctests.Provider.Meta() == nil {
		return nil, errors.New("meta client not inited")
	}

	client := acctests.Provider.Meta().(*transport.Client)
	users, err := client.ReadUsers(context.Background())
	if err != nil {
		return nil, err
	}

	if len(users) == 0 {
		return nil, errors.New("users not found")
	}

	return users[0], nil
}

func testDatasourceTwingateUser(userID string) string {
	return fmt.Sprintf(`
	data "twingate_user" "test" {
	  id = "%s"
	}

	output "my_user_email" {
	  value = data.twingate_user.test.email
	}
	`, userID)
}

func TestAccDatasourceTwingateUser_negative(t *testing.T) {
	t.Parallel()
	t.Run("Test Twingate Datasource : Acc User - does not exists", func(t *testing.T) {
		userID := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("User:%d", acctest.RandInt())))

		resource.Test(t, resource.TestCase{
			ProviderFactories: acctests.ProviderFactories,
			PreCheck: func() {
				acctests.PreCheck(t)
			},
			Steps: []resource.TestStep{
				{
					Config:      testTwingateUserDoesNotExists(userID),
					ExpectError: regexp.MustCompile("Error: failed to read user with id"),
				},
			},
		})
	})
}

func testTwingateUserDoesNotExists(id string) string {
	return fmt.Sprintf(`
	data "twingate_user" "foo" {
	  id = "%s"
	}
	`, id)
}

func TestAccDatasourceTwingateUser_invalidID(t *testing.T) {
	t.Parallel()
	t.Run("Test Twingate Datasource : Acc User - failed parse ID", func(t *testing.T) {
		userID := acctest.RandString(10)

		resource.Test(t, resource.TestCase{
			ProviderFactories: acctests.ProviderFactories,
			PreCheck: func() {
				acctests.PreCheck(t)
			},
			Steps: []resource.TestStep{
				{
					Config:      testTwingateUserDoesNotExists(userID),
					ExpectError: regexp.MustCompile("Unable to parse global ID"),
				},
			},
		})
	})
}

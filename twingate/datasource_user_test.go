package twingate

import (
	"encoding/base64"
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDatasourceTwingateUser_basic(t *testing.T) {
	t.Run("Test Twingate Datasource : Acc User Basic", func(t *testing.T) {
		// TODO: fetch some user for this test
		t.SkipNow()

		const (
			userID    = "VXNlcjoxNDEwMA=="
			userEmail = "eran@twingate.com"
		)

		resource.Test(t, resource.TestCase{
			ProviderFactories: testAccProviderFactories,
			PreCheck:          func() { testAccPreCheck(t) },
			Steps: []resource.TestStep{
				{
					Config: testDatasourceTwingateUser(userID),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckOutput("my_user_email", userEmail),
					),
				},
			},
		})
	})
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
	t.Run("Test Twingate Datasource : Acc User - does not exists", func(t *testing.T) {
		userID := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("User:%d", acctest.RandInt())))

		resource.Test(t, resource.TestCase{
			ProviderFactories: testAccProviderFactories,
			PreCheck: func() {
				testAccPreCheck(t)
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
	t.Run("Test Twingate Datasource : Acc User - failed parse ID", func(t *testing.T) {
		userID := acctest.RandString(10)

		resource.Test(t, resource.TestCase{
			ProviderFactories: testAccProviderFactories,
			PreCheck: func() {
				testAccPreCheck(t)
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

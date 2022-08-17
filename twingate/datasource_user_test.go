package twingate

import (
	"context"
	"encoding/base64"
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDatasourceTwingateUser_basic(t *testing.T) {
	t.Run("Test Twingate Datasource : Acc User Basic", func(t *testing.T) {
		if testAccProvider.Meta() == nil {
			t.Fatal("meta client not inited")
		}

		client := testAccProvider.Meta().(*Client)
		users, err := client.readUsers(context.Background())
		if err != nil {
			t.Fatal(err)
		}

		if len(users) == 0 {
			t.Fatal("users not found")
		}

		user := users[0]

		resource.Test(t, resource.TestCase{
			ProviderFactories: testAccProviderFactories,
			PreCheck:          func() { testAccPreCheck(t) },
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

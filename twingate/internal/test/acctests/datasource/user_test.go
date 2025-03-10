package datasource

import (
	"encoding/base64"
	"fmt"
	"regexp"
	"testing"

	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/test/acctests"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDatasourceTwingateUser_basic(t *testing.T) {
	t.Parallel()

	user, err := acctests.GetTestUser()
	if err != nil {
		t.Skip("can't run test:", err)
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testDatasourceTwingateUser(user.ID),
				Check: acctests.ComposeTestCheckFunc(
					resource.TestCheckOutput("my_user_email_du1", user.Email),
				),
			},
		},
	})
}

func testDatasourceTwingateUser(userID string) string {
	return fmt.Sprintf(`
	data "twingate_user" "test_du1" {
	  id = "%s"
	}

	output "my_user_email_du1" {
	  value = data.twingate_user.test_du1.email
	}
	`, userID)
}

func TestAccDatasourceTwingateUser_negative(t *testing.T) {
	t.Parallel()

	userID := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("User:%d", acctest.RandInt())))

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck: func() {
			acctests.PreCheck(t)
		},
		Steps: []resource.TestStep{
			{
				Config:      testTwingateUserDoesNotExists(userID),
				ExpectError: regexp.MustCompile("failed to read user with id"),
			},
		},
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

	userID := acctest.RandString(10)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck: func() {
			acctests.PreCheck(t)
		},
		Steps: []resource.TestStep{
			{
				Config:      testTwingateUserDoesNotExists(userID),
				ExpectError: regexp.MustCompile("failed to read user with id"),
			},
		},
	})
}

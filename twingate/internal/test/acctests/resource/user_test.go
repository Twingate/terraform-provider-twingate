package resource

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/attr"
	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/model"
	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/provider/resource"
	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/test"
	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/test/acctests"
	sdk "github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
)

func TestAccTwingateUserDeprecateSendInviteTrue(t *testing.T) {
	t.Parallel()

	const terraformResourceName = "test00t"
	theResource := acctests.TerraformUser(terraformResourceName)
	email := test.RandomEmail()

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateUserDestroy,
		Steps: []sdk.TestStep{
			{
				Config: terraformResourceTwingateUserSendInvite(terraformResourceName, email, true),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					sdk.TestCheckResourceAttr(theResource, attr.Email, email),
				),
			},
			{
				PlanOnly: true,
				Config:   terraformResourceTwingateUserSendInvite(terraformResourceName, email, false),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					sdk.TestCheckResourceAttr(theResource, attr.Email, email),
				),
			},
			{
				PlanOnly: true,
				Config:   terraformResourceTwingateUser(terraformResourceName, email),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					sdk.TestCheckResourceAttr(theResource, attr.Email, email),
				),
			},
		},
	})
}

func TestAccTwingateUserDeprecateSendInviteFalse(t *testing.T) {
	t.Parallel()

	const terraformResourceName = "test00f"
	theResource := acctests.TerraformUser(terraformResourceName)
	email := test.RandomEmail()

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateUserDestroy,
		Steps: []sdk.TestStep{
			{
				Config: terraformResourceTwingateUserSendInvite(terraformResourceName, email, false),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					sdk.TestCheckResourceAttr(theResource, attr.Email, email),
				),
			},
			{
				PlanOnly: true,
				Config:   terraformResourceTwingateUserSendInvite(terraformResourceName, email, true),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					sdk.TestCheckResourceAttr(theResource, attr.Email, email),
				),
			},
			{
				PlanOnly: true,
				Config:   terraformResourceTwingateUser(terraformResourceName, email),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					sdk.TestCheckResourceAttr(theResource, attr.Email, email),
				),
			},
		},
	})
}

func TestAccTwingateUserCreateUpdate(t *testing.T) {
	t.Parallel()

	const terraformResourceName = "test001"
	theResource := acctests.TerraformUser(terraformResourceName)
	email := test.RandomEmail()
	firstName := test.RandomName()
	lastName := test.RandomName()
	role := model.UserRoleSupport

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateUserDestroy,
		Steps: []sdk.TestStep{
			{
				Config: terraformResourceTwingateUser(terraformResourceName, email),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					sdk.TestCheckResourceAttr(theResource, attr.Email, email),
				),
			},
			{
				Config: terraformResourceTwingateUserWithFirstName(terraformResourceName, email, firstName),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					sdk.TestCheckResourceAttr(theResource, attr.Email, email),
					sdk.TestCheckResourceAttr(theResource, attr.FirstName, firstName),
				),
			},
			{
				Config: terraformResourceTwingateUserWithLastName(terraformResourceName, email, lastName),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					sdk.TestCheckResourceAttr(theResource, attr.Email, email),
					sdk.TestCheckResourceAttr(theResource, attr.FirstName, firstName),
					sdk.TestCheckResourceAttr(theResource, attr.LastName, lastName),
				),
			},
			{
				Config: terraformResourceTwingateUserWithRole(terraformResourceName, email, role),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					sdk.TestCheckResourceAttr(theResource, attr.Email, email),
					sdk.TestCheckResourceAttr(theResource, attr.FirstName, firstName),
					sdk.TestCheckResourceAttr(theResource, attr.LastName, lastName),
					sdk.TestCheckResourceAttr(theResource, attr.Role, role),
				),
			},
		},
	})
}

func terraformResourceTwingateUserSendInvite(terraformResourceName, email string, sendInvite bool) string {
	return fmt.Sprintf(`
	resource "twingate_user" "%s" {
	  email = "%s"
	  send_invite = %v
	}
	`, terraformResourceName, email, sendInvite)
}

func terraformResourceTwingateUser(terraformResourceName, email string) string {
	return fmt.Sprintf(`
	resource "twingate_user" "%s" {
	  email = "%s"
	}
	`, terraformResourceName, email)
}

func terraformResourceTwingateUserWithFirstName(terraformResourceName, email, firstName string) string {
	return fmt.Sprintf(`
	resource "twingate_user" "%s" {
	  email = "%s"
	  first_name = "%s"
	}
	`, terraformResourceName, email, firstName)
}

func terraformResourceTwingateUserWithLastName(terraformResourceName, email, lastName string) string {
	return fmt.Sprintf(`
	resource "twingate_user" "%s" {
	  email = "%s"
	  last_name = "%s"
	}
	`, terraformResourceName, email, lastName)
}

func terraformResourceTwingateUserWithRole(terraformResourceName, email, role string) string {
	return fmt.Sprintf(`
	resource "twingate_user" "%s" {
	  email = "%s"
	  role = "%s"
	}
	`, terraformResourceName, email, role)
}

func TestAccTwingateUserFullCreate(t *testing.T) {
	t.Parallel()

	const terraformResourceName = "test002"
	theResource := acctests.TerraformUser(terraformResourceName)
	email := test.RandomEmail()
	firstName := test.RandomName()
	lastName := test.RandomName()
	role := test.RandomUserRole()

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateUserDestroy,
		Steps: []sdk.TestStep{
			{
				Config: terraformResourceTwingateUserFull(terraformResourceName, email, firstName, lastName, role),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					sdk.TestCheckResourceAttr(theResource, attr.Email, email),
					sdk.TestCheckResourceAttr(theResource, attr.FirstName, firstName),
					sdk.TestCheckResourceAttr(theResource, attr.LastName, lastName),
					sdk.TestCheckResourceAttr(theResource, attr.Role, role),
				),
			},
		},
	})
}

func terraformResourceTwingateUserFull(terraformResourceName, email, firstName, lastName, role string) string {
	return fmt.Sprintf(`
	resource "twingate_user" "%s" {
	  email = "%s"
	  first_name = "%s"
	  last_name = "%s"
	  role = "%s"
	}
	`, terraformResourceName, email, firstName, lastName, role)
}

func TestAccTwingateUserReCreation(t *testing.T) {
	t.Parallel()

	const terraformResourceName = "test003"
	theResource := acctests.TerraformUser(terraformResourceName)
	email1 := test.RandomEmail()
	email2 := test.RandomEmail()

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateUserDestroy,
		Steps: []sdk.TestStep{
			{
				Config: terraformResourceTwingateUser(terraformResourceName, email1),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					sdk.TestCheckResourceAttr(theResource, attr.Email, email1),
				),
			},
			{
				Config: terraformResourceTwingateUser(terraformResourceName, email2),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					sdk.TestCheckResourceAttr(theResource, attr.Email, email2),
				),
			},
		},
	})
}

func TestAccTwingateUserUpdateState(t *testing.T) {
	t.Parallel()

	const terraformResourceName = "test004"
	theResource := acctests.TerraformUser(terraformResourceName)
	email := test.RandomEmail()

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateUserDestroy,
		Steps: []sdk.TestStep{
			{
				Config: terraformResourceTwingateUser(terraformResourceName, email),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					sdk.TestCheckResourceAttr(theResource, attr.Email, email),
				),
			},
			{
				Config:      terraformResourceTwingateUserDisabled(terraformResourceName, email),
				ExpectError: regexp.MustCompile(`User in PENDING state`),
			},
		},
	})
}

func terraformResourceTwingateUserDisabled(terraformResourceName, email string) string {
	return fmt.Sprintf(`
	resource "twingate_user" "%s" {
	  email = "%s"
	  is_active = false
	}
	`, terraformResourceName, email)
}

func TestAccTwingateUserDelete(t *testing.T) {
	t.Parallel()

	const terraformResourceName = "test005"
	theResource := acctests.TerraformUser(terraformResourceName)
	userEmail := test.RandomEmail()

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateUserDestroy,
		Steps: []sdk.TestStep{
			{
				Config:  terraformResourceTwingateUser(terraformResourceName, userEmail),
				Destroy: true,
			},
			{
				Config: terraformResourceTwingateUser(terraformResourceName, userEmail),
				ConfigPlanChecks: sdk.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(theResource, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

func TestAccTwingateUserReCreateAfterDeletion(t *testing.T) {
	t.Parallel()

	const terraformResourceName = "test006"
	theResource := acctests.TerraformUser(terraformResourceName)
	email := test.RandomEmail()

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateUserDestroy,
		Steps: []sdk.TestStep{
			{
				Config: terraformResourceTwingateUser(terraformResourceName, email),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
					acctests.DeleteTwingateResource(theResource, resource.TwingateUser),
				),
				ExpectNonEmptyPlan: true,
			},
			{
				Config: terraformResourceTwingateUser(terraformResourceName, email),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
				),
			},
		},
	})
}

func TestAccTwingateUserCreateWithUnknownRole(t *testing.T) {
	t.Parallel()

	const terraformResourceName = "test007"

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateUserDestroy,
		Steps: []sdk.TestStep{
			{
				Config:      terraformResourceTwingateUserWithRole(terraformResourceName, test.RandomEmail(), "UnknownRole"),
				ExpectError: regexp.MustCompile(`Attribute role value must be one of`),
			},
		},
	})
}

func TestAccTwingateUserCreateWithoutEmail(t *testing.T) {
	t.Parallel()

	const terraformResourceName = "test008"

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateUserDestroy,
		Steps: []sdk.TestStep{
			{
				Config:      terraformResourceTwingateUserWithoutEmail(terraformResourceName),
				ExpectError: regexp.MustCompile("Error: Missing required argument"),
			},
		},
	})
}

func terraformResourceTwingateUserWithoutEmail(terraformResourceName string) string {
	return fmt.Sprintf(`
	resource "twingate_user" "%s" {
	}
	`, terraformResourceName)
}

func genNewUsers(resourcePrefix string, count int) ([]string, []string) {
	users := make([]string, 0, count)
	userIDs := make([]string, 0, count)

	for i := 0; i < count; i++ {
		resourceName := fmt.Sprintf("%s_%d", resourcePrefix, i+1)
		users = append(users, terraformResourceTwingateUser(resourceName, test.RandomEmail()))
		userIDs = append(userIDs, fmt.Sprintf("twingate_user.%s.id", resourceName))
	}

	return users, userIDs
}

func TestAccTwingateUserImport(t *testing.T) {
	t.Parallel()

	const terraformResourceName = "test009"
	theResource := acctests.TerraformUser(terraformResourceName)
	email := test.RandomEmail()
	firstName := test.RandomName()
	lastName := test.RandomName()
	role := test.RandomUserRole()

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		CheckDestroy:             acctests.CheckTwingateUserDestroy,
		Steps: []sdk.TestStep{
			{
				Config: terraformResourceTwingateUserFull(terraformResourceName, email, firstName, lastName, role),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(theResource),
				),
			},
			{
				ImportState:  true,
				ResourceName: theResource,
				ImportStateCheck: acctests.CheckImportState(map[string]string{
					attr.Email:     email,
					attr.FirstName: firstName,
					attr.LastName:  lastName,
					attr.Role:      role,
					attr.Type:      model.UserTypeManual,
					attr.IsActive:  "true",
				}),
			},
		},
	})
}

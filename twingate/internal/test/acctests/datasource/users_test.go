package datasource

import (
	"fmt"
	"strings"
	"testing"

	"github.com/Twingate/terraform-provider-twingate/twingate/internal/attr"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/test"
	"github.com/Twingate/terraform-provider-twingate/twingate/internal/test/acctests"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccDatasourceTwingateUsers_basic(t *testing.T) {
	t.Run("Test Twingate Datasource : Acc Users Basic", func(t *testing.T) {
		acctests.SetPageLimit(1)
		resource.Test(t, resource.TestCase{
			ProtoV6ProviderFactories: acctests.ProviderFactories,
			PreCheck:                 func() { acctests.PreCheck(t) },
			Steps: []resource.TestStep{
				{
					Config: testDatasourceTwingateUsers(),
					Check: acctests.ComposeTestCheckFunc(
						testCheckResourceAttrNotEqual("data.twingate_users.all", attr.Len(attr.Users), "0"),
					),
				},
			},
		})
	})
}

func testDatasourceTwingateUsers() string {
	return `
	data "twingate_users" "all" {}
	`
}

func testCheckResourceAttrNotEqual(name, key, value string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		ms := s.RootModule()

		res, ok := ms.Resources[name]
		if !ok || res == nil || res.Primary == nil {
			return fmt.Errorf("resource '%s' not found", name)
		}

		actual, ok := res.Primary.Attributes[key]
		if !ok {
			return fmt.Errorf("attribute '%s' not found", key)
		}

		if actual == value {
			return fmt.Errorf("expected not equal value '%s', but got equal", value)
		}

		return nil
	}
}

func join(configs ...string) string {
	return strings.Join(configs, "\n")
}

func TestAccDatasourceTwingateUsers_filterByEmail(t *testing.T) {
	t.Parallel()

	resourceName := test.RandomName()
	datasourceName := test.RandomName()
	prefix := test.TerraformRandName("email")
	email := prefix + "_" + test.RandomEmail()
	theDatasource := acctests.TerraformDatasourceUsers(datasourceName)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: join(
					terraformResourceTwingateUser(resourceName, email),
					terraformDatasourceUsersByEmail(datasourceName, "", email, resourceName),
				),
				Check: acctests.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(theDatasource, attr.Len(attr.Users), "1"),
					resource.TestCheckResourceAttr(theDatasource, attr.Path(attr.Users, attr.Email), email),
				),
			},
		},
	})
}

func TestAccDatasourceTwingateUsers_filterByEmailPrefix(t *testing.T) {
	t.Parallel()

	resourceName := test.RandomName()
	datasourceName := test.RandomName()
	prefix := test.TerraformRandName("email_prefix")
	email := prefix + "_" + test.RandomEmail()
	theDatasource := acctests.TerraformDatasourceUsers(datasourceName)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: join(
					terraformResourceTwingateUser(resourceName, email),
					terraformDatasourceUsersByEmail(datasourceName, attr.FilterByPrefix, prefix, resourceName),
				),
				Check: acctests.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(theDatasource, attr.Len(attr.Users), "1"),
					resource.TestCheckResourceAttr(theDatasource, attr.Path(attr.Users, attr.Email), email),
				),
			},
		},
	})
}

func TestAccDatasourceTwingateUsers_filterByEmailSuffix(t *testing.T) {
	t.Parallel()

	resourceName := test.RandomName()
	datasourceName := test.RandomName()
	const suffix = "suf"
	email := test.RandomEmail() + "." + suffix
	theDatasource := acctests.TerraformDatasourceUsers(datasourceName)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: join(
					terraformResourceTwingateUser(resourceName, email),
					terraformDatasourceUsersByEmail(datasourceName, attr.FilterBySuffix, suffix, resourceName),
				),
				Check: acctests.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(theDatasource, attr.Len(attr.Users), "1"),
					resource.TestCheckResourceAttr(theDatasource, attr.Path(attr.Users, attr.Email), email),
				),
			},
		},
	})
}

func TestAccDatasourceTwingateUsers_filterByEmailContains(t *testing.T) {
	t.Parallel()

	resourceName := test.RandomName()
	datasourceName := test.RandomName()

	val := acctest.RandString(6)
	email := test.TerraformRandName(val) + "_" + test.RandomEmail()
	theDatasource := acctests.TerraformDatasourceUsers(datasourceName)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: join(
					terraformResourceTwingateUser(resourceName, email),
					terraformDatasourceUsersByEmail(datasourceName, attr.FilterByContains, val, resourceName),
				),
				Check: acctests.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(theDatasource, attr.Len(attr.Users), "1"),
					resource.TestCheckResourceAttr(theDatasource, attr.Path(attr.Users, attr.Email), email),
				),
			},
		},
	})
}

func TestAccDatasourceTwingateUsers_filterByEmailRegexp(t *testing.T) {
	t.Parallel()

	resourceName := test.RandomName()
	datasourceName := test.RandomName()

	prefix := acctest.RandString(6)
	email := test.TerraformRandName(prefix) + "_email_by_regexp_" + test.RandomEmail()
	theDatasource := acctests.TerraformDatasourceUsers(datasourceName)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: join(
					terraformResourceTwingateUser(resourceName, email),
					terraformDatasourceUsersByEmail(datasourceName, attr.FilterByRegexp, prefix+".*_regexp_.*", resourceName),
				),
				Check: acctests.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(theDatasource, attr.Len(attr.Users), "1"),
					resource.TestCheckResourceAttr(theDatasource, attr.Path(attr.Users, attr.Email), email),
				),
			},
		},
	})
}

func terraformResourceTwingateUser(terraformResourceName, email string) string {
	return fmt.Sprintf(`
	resource "twingate_user" "%s" {
	  email = "%s"
	  send_invite = false
	}
	`, terraformResourceName, email)
}

func terraformDatasourceUsersByEmail(datasourceName, filter, email, resourceName string) string {
	return fmt.Sprintf(`
	data "twingate_users" "%[1]s" {
	  email%[2]s = "%[3]s"

	  depends_on = [%[4]s]
	}
`, datasourceName, filter, email, acctests.TerraformUser(resourceName))
}

func TestAccDatasourceTwingateUsers_filterByEmailAndFirstName(t *testing.T) {
	t.Parallel()

	resourceName := test.RandomName()
	prefix := test.TerraformRandName("orange")
	email := prefix + "_" + test.RandomEmail()
	firstName := prefix + "_" + test.RandomName()
	const theDatasource = "data.twingate_users.filter_by_email_and_first_name_prefix"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: join(
					terraformResourceTwingateUserWithFirstName(resourceName, email, firstName),
					terraformDatasourceUsersByEmailAndFirstNamePrefix(prefix, resourceName),
				),
				Check: acctests.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(theDatasource, attr.Len(attr.Users), "1"),
					resource.TestCheckResourceAttr(theDatasource, attr.Path(attr.Users, attr.Email), email),
					resource.TestCheckResourceAttr(theDatasource, attr.Path(attr.Users, attr.FirstName), firstName),
				),
			},
		},
	})
}

func terraformResourceTwingateUserWithFirstName(resourceName, email, firstName string) string {
	return fmt.Sprintf(`
	resource "twingate_user" "%s" {
	  email = "%s"
	  first_name = "%s"
	  send_invite = false
	}
	`, resourceName, email, firstName)
}

func terraformDatasourceUsersByEmailAndFirstNamePrefix(prefix, resourceName string) string {
	return fmt.Sprintf(`
	data "twingate_users" "filter_by_email_and_first_name_prefix" {
	  email_prefix = "%[1]s"
	  first_name_prefix = "%[1]s"

	  depends_on = [twingate_user.%[2]s]
	}
`, prefix, resourceName)
}

func TestAccDatasourceTwingateUsers_filterByEmailFirstNameAndLastName(t *testing.T) {
	t.Parallel()

	resourceName := test.RandomName()
	prefix := test.TerraformRandName("yellow")
	email := prefix + "_" + test.RandomEmail()
	firstName := prefix + "_" + test.RandomName()
	lastName := prefix + "_" + test.RandomName()
	const theDatasource = "data.twingate_users.filter_by_email_and_first_name_and_last_name_prefix"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: join(
					terraformResourceTwingateUserWithFirstNameAndLastName(resourceName, email, firstName, lastName),
					terraformDatasourceUsersByEmailAndFirstNameAndLastNamePrefix(prefix, resourceName),
				),
				Check: acctests.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(theDatasource, attr.Len(attr.Users), "1"),
					resource.TestCheckResourceAttr(theDatasource, attr.Path(attr.Users, attr.Email), email),
					resource.TestCheckResourceAttr(theDatasource, attr.Path(attr.Users, attr.FirstName), firstName),
					resource.TestCheckResourceAttr(theDatasource, attr.Path(attr.Users, attr.LastName), lastName),
				),
			},
		},
	})
}

func terraformResourceTwingateUserWithFirstNameAndLastName(resourceName, email, firstName, lastName string) string {
	return fmt.Sprintf(`
	resource "twingate_user" "%s" {
	  email = "%s"
	  first_name = "%s"
	  last_name = "%s"
	  send_invite = false
	}
	`, resourceName, email, firstName, lastName)
}

func terraformDatasourceUsersByEmailAndFirstNameAndLastNamePrefix(prefix, resourceName string) string {
	return fmt.Sprintf(`
	data "twingate_users" "filter_by_email_and_first_name_and_last_name_prefix" {
	  email_prefix = "%[1]s"
	  first_name_prefix = "%[1]s"
	  last_name_prefix = "%[1]s"

	  depends_on = [twingate_user.%[2]s]
	}
`, prefix, resourceName)
}

func TestAccDatasourceTwingateUsers_filterByEmailFirstNameLastNameAndRole(t *testing.T) {
	t.Parallel()

	resourceName := test.RandomName()
	prefix := test.TerraformRandName("tree")
	email := prefix + "_" + test.RandomEmail()
	firstName := prefix + "_" + test.RandomName()
	lastName := prefix + "_" + test.RandomName()
	const theDatasource = "data.twingate_users.filter_by_email_first-name_last-name_prefix_and_role"
	const role = "DEVOPS"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: join(
					terraformResourceTwingateUserWithFirstNameLastNameAndRole(resourceName, email, firstName, lastName, role),
					terraformDatasourceUsersByEmailAndFirstNameLastNamePrefixAndRole(prefix, resourceName, role),
				),
				Check: acctests.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(theDatasource, attr.Len(attr.Users), "1"),
					resource.TestCheckResourceAttr(theDatasource, attr.Path(attr.Users, attr.Email), email),
					resource.TestCheckResourceAttr(theDatasource, attr.Path(attr.Users, attr.FirstName), firstName),
					resource.TestCheckResourceAttr(theDatasource, attr.Path(attr.Users, attr.LastName), lastName),
					resource.TestCheckResourceAttr(theDatasource, attr.Path(attr.Users, attr.Role), role),
				),
			},
		},
	})
}

func terraformResourceTwingateUserWithFirstNameLastNameAndRole(resourceName, email, firstName, lastName, role string) string {
	return fmt.Sprintf(`
	resource "twingate_user" "%s" {
	  email = "%s"
	  first_name = "%s"
	  last_name = "%s"
	  role = "%s"
	  send_invite = false
	}
	`, resourceName, email, firstName, lastName, role)
}

func terraformDatasourceUsersByEmailAndFirstNameLastNamePrefixAndRole(prefix, resourceName, role string) string {
	return fmt.Sprintf(`
	data "twingate_users" "filter_by_email_first-name_last-name_prefix_and_role" {
	  email_prefix = "%[1]s"
	  first_name_prefix = "%[1]s"
	  last_name_prefix = "%[1]s"
	  roles = ["%[2]s"]

	  depends_on = [twingate_user.%[3]s]
	}
`, prefix, role, resourceName)
}

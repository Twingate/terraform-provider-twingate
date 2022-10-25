package twingate

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/twingate/go-graphql-client"
)

const groupResource = "twingate_group.test"

func TestAccTwingateGroupCreateUpdate(t *testing.T) {
	t.Run("Test Twingate Resource : Acc Group Create/Update", func(t *testing.T) {

		groupNameBefore := getRandomName()
		groupNameAfter := getRandomName()

		const theResource = "twingate_group.test001"

		resource.ParallelTest(t, resource.TestCase{
			ProviderFactories: testAccProviderFactories,
			PreCheck:          func() { testAccPreCheck(t) },
			CheckDestroy:      testAccCheckTwingateGroupDestroy,
			Steps: []resource.TestStep{
				{
					Config: createGroup001(groupNameBefore),
					Check: ComposeTestCheckFunc(
						testAccCheckTwingateGroupExists(theResource),
						resource.TestCheckResourceAttr(theResource, "name", groupNameBefore),
					),
				},
				{
					Config: createGroup001(groupNameAfter),
					Check: ComposeTestCheckFunc(
						testAccCheckTwingateGroupExists(theResource),
						resource.TestCheckResourceAttr(theResource, "name", groupNameAfter),
					),
				},
			},
		})
	})
}

func createGroup001(name string) string {
	return fmt.Sprintf(`
	resource "twingate_group" "test001" {
	  name = "%s"
	}
	`, name)
}

func TestAccTwingateGroupDeleteNonExisting(t *testing.T) {
	t.Run("Test Twingate Resource : Acc Group Delete NonExisting", func(t *testing.T) {

		groupNameBefore := getRandomName()

		resource.ParallelTest(t, resource.TestCase{
			ProviderFactories: testAccProviderFactories,
			PreCheck:          func() { testAccPreCheck(t) },
			CheckDestroy:      testAccCheckTwingateGroupDestroy,
			Steps: []resource.TestStep{
				{
					Config:  createGroup002(groupNameBefore),
					Destroy: true,
					Check: ComposeTestCheckFunc(
						testAccCheckTwingateGroupDoesNotExists("twingate_group.test002"),
					),
				},
			},
		})
	})
}

func createGroup002(name string) string {
	return fmt.Sprintf(`
	resource "twingate_group" "test002" {
	  name = "%s"
	}
	`, name)
}

func testAccCheckTwingateGroupDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "twingate_group" {
			continue
		}

		groupId := rs.Primary.ID

		err := client.deleteGroup(context.Background(), groupId)
		if err == nil {
			return fmt.Errorf("Group with ID %s still present : ", groupId)
		}
	}

	return nil
}

func testAccCheckTwingateGroupExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]

		if !ok {
			return fmt.Errorf("Not found: %s ", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No GroupID set ")
		}

		return nil
	}
}

func testAccCheckTwingateGroupDoesNotExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		_ = rs
		if !ok {
			return nil
		}

		return fmt.Errorf("this resource should not be here: %s ", resourceName)
	}
}

func TestAccTwingateGroupReCreateAfterDeletion(t *testing.T) {
	t.Run("Test Twingate Resource : Acc Group Create After Deletion", func(t *testing.T) {
		groupName := getRandomName()

		const theResource = "twingate_group.test003"

		resource.ParallelTest(t, resource.TestCase{
			ProviderFactories: testAccProviderFactories,
			PreCheck:          func() { testAccPreCheck(t) },
			CheckDestroy:      testAccCheckTwingateGroupDestroy,
			Steps: []resource.TestStep{
				{
					Config: createGroup003(groupName),
					Check: ComposeTestCheckFunc(
						testAccCheckTwingateGroupExists(theResource),
						deleteTwingateResource(theResource, groupResourceName),
					),
					ExpectNonEmptyPlan: true,
				},
				{
					Config: createGroup003(groupName),
					Check: ComposeTestCheckFunc(
						testAccCheckTwingateGroupExists(theResource),
					),
				},
			},
		})
	})
}

func createGroup003(name string) string {
	return fmt.Sprintf(`
	resource "twingate_group" "test003" {
	  name = "%s"
	}
	`, name)
}

func TestAccTwingateGroupCreateWithUsersAndResources(t *testing.T) {
	t.Run("Test Twingate Resource : Acc Group Create With Users And Resources", func(t *testing.T) {
		groupName := getRandomName()

		user, err := getTestUser()
		if err != nil {
			t.Skip("can't run test:", err)
		}

		resID, err := getTestResourceID()
		if err != nil {
			t.Skip("can't run test:", err)
		}

		const theResource = "twingate_group.test004"

		resource.ParallelTest(t, resource.TestCase{
			ProviderFactories: testAccProviderFactories,
			PreCheck:          func() { testAccPreCheck(t) },
			CheckDestroy:      testAccCheckTwingateGroupDestroy,
			Steps: []resource.TestStep{
				{
					Config: createGroup004WithUser(groupName, user.ID),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckTwingateGroupExists(theResource),
						resource.TestCheckResourceAttr(theResource, "name", groupName),
						resource.TestCheckResourceAttr(theResource, "users.#", "1"),
						resource.TestCheckResourceAttr(theResource, "users.0", user.ID),
						resource.TestCheckResourceAttr(theResource, "resources.#", "0"),
					),
				},
				{
					Config: createGroup004WithResource(groupName, resID),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckTwingateGroupExists(theResource),
						resource.TestCheckResourceAttr(theResource, "name", groupName),
						resource.TestCheckResourceAttr(theResource, "users.#", "0"),
						resource.TestCheckResourceAttr(theResource, "resources.#", "1"),
						resource.TestCheckResourceAttr(theResource, "resources.0", resID),
					),
				},
			},
		})
	})
}

func createGroup004WithUser(name, userID string) string {
	return fmt.Sprintf(`
	resource "twingate_group" "test004" {
	  name = "%s"
	  users = ["%s"]
	}
	`, name, userID)
}

func createGroup004WithResource(name, resourceID string) string {
	return fmt.Sprintf(`
	resource "twingate_group" "test004" {
	  name = "%s"
	  resources = ["%s"]
	}
	`, name, resourceID)
}

func getTestResourceID() (string, error) {
	if testAccProvider.Meta() == nil {
		return "", errors.New("meta client not inited")
	}

	client := testAccProvider.Meta().(*Client)
	resources, err := client.readResources(context.Background())
	if err != nil {
		return "", err
	}

	if len(resources) == 0 {
		return "", errors.New("resources not found")
	}

	return resources[0].Node.ID.(string), nil
}

func TestToStringID(t *testing.T) {
	cases := []struct {
		input    []graphql.ID
		expected []string
	}{
		{
			input:    nil,
			expected: []string{},
		},
		{
			input:    []graphql.ID{},
			expected: []string{},
		},
		{
			input:    []graphql.ID{"id"},
			expected: []string{"id"},
		},
	}

	for n, c := range cases {
		t.Run(fmt.Sprintf("case #%d", n), func(t *testing.T) {
			actual := toStringIDs(c.input)
			assert.Equal(t, c.expected, actual)
		})
	}
}

func TestConvertTerraformListToGraphqlIDs(t *testing.T) {
	var s *schema.Set
	cases := []struct {
		input    interface{}
		expected []graphql.ID
	}{
		{
			input:    nil,
			expected: []graphql.ID{},
		},
		{
			input:    interface{}("hello"),
			expected: []graphql.ID{},
		},
		{
			input:    interface{}(&schema.Set{}),
			expected: []graphql.ID{},
		},
		{
			input:    interface{}(s),
			expected: []graphql.ID{},
		},
		{
			input: interface{}(schema.NewSet(schema.HashString, []interface{}{
				"id",
			})),
			expected: []graphql.ID{"id"},
		},
	}

	for n, c := range cases {
		t.Run(fmt.Sprintf("case #%d", n), func(t *testing.T) {
			actual := convertTerraformListToGraphqlIDs(c.input)
			assert.Equal(t, c.expected, actual)
		})
	}
}

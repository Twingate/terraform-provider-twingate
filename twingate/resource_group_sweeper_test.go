package twingate

import (
	"context"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hasura/go-graphql-client"
)

func init() {
	resource.AddTestSweepers("twingate_group", &resource.Sweeper{
		Name: "twingate_group",
		F:    testSweepTwingateGroup,
	})
}

func testSweepTwingateGroup(tenant string) error {
	resourceName := "twingate_group"
	log.Printf("\"[INFO][SWEEPER_LOG] Starting sweeper for %s\"", resourceName)
	client, err := sharedClient(tenant)
	if err != nil {
		log.Printf("[ERROR][SWEEPER_LOG] error getting client: %s", err)
		return err
	}

	ctx := context.Background()

	resources, err := client.readGroups(ctx)
	if err != nil {
		log.Printf("[INFO][SWEEPER_LOG] Nothing found in response: %s", resourceName)
		return nil
	}

	if len(resources) == 0 {
		log.Printf("[INFO][SWEEPER_LOG] List %s is empty", resourceName)
		return nil
	}

	var testResources = make([]graphql.ID, 0)

	for _, elem := range resources {
		if strings.HasPrefix(string(elem.Name), testPrefixName) {
			testResources = append(testResources, elem.ID)
		}
	}

	if len(testResources) == 0 {
		log.Printf("[INFO][SWEEPER_LOG] List with test networks %s is empty", resourceName)
		return nil
	}

	for _, i := range testResources {
		if i == "" {
			log.Printf("[INFO][SWEEPER_LOG] %s: %s name was empty value", resourceName, i)
			return nil
		}
		err = client.deleteGroup(ctx, i.(string))
		if err != nil {
			log.Printf("[INFO][SWEEPER_LOG] %s cannot be deleted, error: %s", i, err)
			return nil
		}
	}

	return nil
}

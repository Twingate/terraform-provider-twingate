package sweepers

import (
	"context"
	"log"
	"os"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	testPrefixName = "tf-acc"
	uniqueEnvKey   = "TEST_UNIQUE_VALUE"
)

func getRandomConnectorName() string {
	const maxLength = 30
	name := getTestPrefix(acctest.RandString(4))
	if len(name) > maxLength {
		name = name[:maxLength]
	}

	return name
}

func getRandomResourceName() string {
	return getRandomName("resource")
}

func getRandomGroupName() string {
	return getRandomName("group")
}

func getRandomName(names ...string) string {
	return acctest.RandomWithPrefix(getTestPrefix(names...))
}

func getTestPrefix(names ...string) string {
	uniqueVal := os.Getenv(uniqueEnvKey)
	uniqueVal = strings.ReplaceAll(uniqueVal, ".", "")
	uniqueVal = strings.ReplaceAll(uniqueVal, "*", "")

	keys := filterStringValues(
		append([]string{testPrefixName, uniqueVal}, names...),
		func(val string) bool {
			return strings.TrimSpace(val) != ""
		},
	)

	return strings.Join(keys, "-")
}

func filterStringValues(values []string, ok func(val string) bool) []string {
	result := make([]string, 0, len(values))
	for _, val := range values {
		if ok(val) {
			result = append(result, val)
		}
	}

	return result
}

func init() {
	resource.AddTestSweepers("twingate_connector", &resource.Sweeper{
		Name: "twingate_connector",
		F:    testSweepTwingateConnector,
	})
}

func testSweepTwingateConnector(tenant string) error {
	resourceName := "twingate_connector"
	log.Printf("[INFO][SWEEPER_LOG] Starting sweeper for %s", resourceName)
	client, err := sharedClient(tenant)
	if err != nil {
		log.Printf("[ERROR][SWEEPER_LOG] error getting client: %s", err)
		return err
	}

	ctx := context.Background()

	connectorMap, err := client.readConnectors(ctx)
	if err != nil {
		log.Printf("[INFO][SWEEPER_LOG] Nothing found in response: %s", resourceName)
		return nil
	}

	if len(connectorMap) == 0 {
		log.Printf("[INFO][SWEEPER_LOG] List %s is empty", resourceName)
		return nil
	}

	var testConnectors = make([]string, 0)

	for _, elem := range connectorMap {
		if strings.HasPrefix(elem.Name, testPrefixName) {
			testConnectors = append(testConnectors, elem.ID)
		}
	}

	if len(testConnectors) == 0 {
		log.Printf("[INFO][SWEEPER_LOG] List with test connectors %s is empty", resourceName)
		return nil
	}

	for _, i := range testConnectors {
		if i == "" {
			log.Printf("[INFO][SWEEPER_LOG] %s: %s name was empty value", resourceName, i)
			continue
		}
		err = client.deleteConnector(ctx, i)
		if err != nil {
			log.Printf("[INFO][SWEEPER_LOG] %s cannot be deleted, error: %s", i, err)
			continue
		}
	}

	return nil
}

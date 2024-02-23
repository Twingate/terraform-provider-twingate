package test

import (
	"crypto/sha1"
	"fmt"
	"os"
	"strings"

	"github.com/Twingate/terraform-provider-twingate/twingate/internal/model"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
)

const (
	prefixName   = "tf-acc"
	envUniqueVal = "TEST_UNIQUE_VALUE"
)

func RandomConnectorName() string {
	const maxLength = 30

	hash := fmt.Sprintf("%x", sha1.Sum([]byte(Prefix())))
	name := fmt.Sprintf("%s-%s-", prefixName, hash[:15])
	name += acctest.RandString(maxLength - len(name))

	return name
}

func RandomResourceName() string {
	return RandomName("resource")
}

func RandomGroupName() string {
	return RandomName("group")
}

func RandomName(names ...string) string {
	return acctest.RandomWithPrefix(Prefix(names...))
}

func RandomEmail() string {
	const (
		nameLen   = 6
		domainLen = 4
	)

	return fmt.Sprintf("%s_%s@%s.com", Prefix(), acctest.RandString(nameLen), acctest.RandString(domainLen))
}

func RandomUserRole() string {
	return model.UserRoles[acctest.RandIntRange(0, len(model.UserRoles)-1)]
}

func Prefix(names ...string) string {
	uniqueVal := os.Getenv(envUniqueVal)
	uniqueVal = strings.ReplaceAll(uniqueVal, ".", "")
	uniqueVal = strings.ReplaceAll(uniqueVal, "*", "")

	keys := filterStringValues(
		append([]string{prefixName, uniqueVal}, names...),
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

func TerraformRandName(name string) string {
	return fmt.Sprintf("%s_%02d", name, acctest.RandInt())
}

package resource

import (
	"fmt"

	"github.com/Twingate/terraform-provider-twingate/twingate/internal/utils"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func ErrAttributeSet(err error, attribute string) diag.Diagnostics {
	return diag.FromErr(fmt.Errorf("error setting %s: %w ", attribute, err))
}

func castToStrings(a, b interface{}) (string, string) {
	return a.(string), b.(string)
}

func convertIDs(data interface{}) []string {
	return utils.Map[interface{}, string](
		data.(*schema.Set).List(),
		func(elem interface{}) string {
			return elem.(string)
		},
	)
}

package customvalidator

import (
	"context"
	"fmt"
	"net"

	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatordiag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var _ validator.String = addressValidator{}

type addressValidator struct{}

func (v addressValidator) Description(_ context.Context) string {
	return `string must be a valid IP or FQDN (like server.example.com)`
}

func (v addressValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v addressValidator) validate(value string) error {
	if hostnameRgxp.MatchString(value) {
		return nil
	}

	if net.ParseIP(value) == nil {
		return fmt.Errorf("invalid IP address: %q", value) //nolint:err113
	}

	return nil
}

func (v addressValidator) ValidateString(ctx context.Context, request validator.StringRequest, response *validator.StringResponse) {
	if request.ConfigValue.IsNull() || request.ConfigValue.IsUnknown() {
		return
	}

	if err := v.validate(request.ConfigValue.ValueString()); err != nil {
		response.Diagnostics.Append(validatordiag.InvalidAttributeValueDiagnostic(
			request.Path,
			v.Description(ctx),
			err.Error(),
		))
	}
}

// Address returns a validator that ensures a string is a valid IP/FQDN address.
func Address() validator.String {
	return addressValidator{}
}

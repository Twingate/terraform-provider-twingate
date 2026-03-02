package customvalidator

import (
	"context"
	"fmt"
	"net"
	"regexp"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatordiag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var _ validator.String = addressValidator{}

var hostnameRgxp = regexp.MustCompile(`^(?:[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?\.)+[a-zA-Z]{2,63}$`)

type addressValidator struct{}

func (v addressValidator) Description(_ context.Context) string {
	return `string must be a valid "host:port" address (e.g. "10.0.0.1:8080")`
}

func (v addressValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v addressValidator) validate(value string) error {
	host, portStr, err := net.SplitHostPort(value)
	if err != nil {
		return fmt.Errorf(`invalid format, expected "host:port": %w`, err)
	}

	port, err := strconv.Atoi(portStr)
	if err != nil || port < 1 || port > 65535 {
		return fmt.Errorf("invalid port %q: must be an integer between 1 and 65535", portStr) //nolint:err113
	}

	if hostnameRgxp.MatchString(host) {
		return nil
	}

	if net.ParseIP(host) == nil {
		return fmt.Errorf("invalid IP address: %q", host) //nolint:err113
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

// Address returns a validator that ensures a string is a valid "IP:port" address.
func Address() validator.String {
	return addressValidator{}
}

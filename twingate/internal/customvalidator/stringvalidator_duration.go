package customvalidator

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"

	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatordiag"
	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatorfuncerr"
)

var _ validator.String = durationValidator{}
var _ function.StringParameterValidator = durationValidator{}

type durationValidator struct {
}

func (validator durationValidator) invalidUsageMessage() string {
	return "string must be a valid duration"
}

func (validator durationValidator) Description(_ context.Context) string {
	return validator.invalidUsageMessage()
}

func (validator durationValidator) MarkdownDescription(ctx context.Context) string {
	return validator.Description(ctx)
}

func (v durationValidator) validate(value string) error {
	duration, err := time.ParseDuration(value)
	if err != nil {
		return fmt.Errorf("failed to parse Duration %v", err)
	}

	if duration < 0 {
		return fmt.Errorf("got negative Duration %v", duration.String())
	}

	return nil
}

func (v durationValidator) ValidateString(ctx context.Context, request validator.StringRequest, response *validator.StringResponse) {
	if request.ConfigValue.IsNull() || request.ConfigValue.IsUnknown() {
		return
	}

	value := request.ConfigValue.ValueString()

	if err := v.validate(value); err != nil {
		response.Diagnostics.Append(validatordiag.InvalidAttributeValueDiagnostic(
			request.Path,
			v.Description(ctx),
			err.Error(),
		))
	}
}

func (v durationValidator) ValidateParameterString(ctx context.Context, request function.StringParameterValidatorRequest, response *function.StringParameterValidatorResponse) {
	if request.Value.IsNull() || request.Value.IsUnknown() {
		return
	}

	value := request.Value.ValueString()

	if err := v.validate(value); err != nil {
		response.Error = validatorfuncerr.InvalidParameterValueFuncError(
			request.ArgumentPosition,
			v.Description(ctx),
			err.Error(),
		)
	}
}

// Duration returns a validator which ensures that duration string configured correctly
func Duration() durationValidator {
	return durationValidator{}
}

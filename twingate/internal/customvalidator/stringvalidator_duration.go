package customvalidator

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"

	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatordiag"
	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatorfuncerr"

	"github.com/Twingate/terraform-provider-twingate/v3/twingate/internal/utils"
)

var (
	_ validator.String                  = durationValidator{}
	_ function.StringParameterValidator = durationValidator{}

	ErrLessThenMinDuration = errors.New("minimum duration is 1 hour")
	ErrExceedsMaxDuration  = errors.New("maximum duration is 365 days")
)

const (
	minDuration time.Duration = time.Hour
	maxDuration time.Duration = time.Hour * 24 * 365
)

type durationValidator struct{}

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
	duration, err := utils.ParseDurationWithDays(value)
	if err != nil {
		return fmt.Errorf("failed to parse duration %w", err)
	}

	if duration < minDuration {
		return ErrLessThenMinDuration
	}

	if duration > maxDuration {
		return ErrExceedsMaxDuration
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

// Duration returns a validator which ensures that duration string configured correctly.
func Duration() durationValidator {
	return durationValidator{}
}

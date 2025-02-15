package resource

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSanitizeName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"  connector   name  ", "connector name"},                   // Leading, trailing, and extra spaces
		{"connector name", "connector name"},                         // No extra spaces
		{"    connector    multi   space ", "connector multi space"}, // Multi spaces
		{"", ""},    // Empty input
		{"   ", ""}, // Only spaces
	}

	for _, test := range tests {
		result := sanitizeName(test.input)
		if result != test.expected {
			t.Errorf("sanitizeName(%q) = %q; want %q", test.input, result, test.expected)
		}
	}
}

func TestSanitizedLengthValidator_ValidateString(t *testing.T) {
	tests := []struct {
		name           string
		input          types.String
		minLen         int
		expectedErrors bool
	}{
		{"Valid Input", types.StringValue("valid name"), 5, false},
		{"Invalid Input - Too Short", types.StringValue("abc"), 5, true},
		{"Valid Input After Sanitize", types.StringValue("    valid name   "), 5, false},
		{"Invalid After Sanitize", types.StringValue(" a "), 5, true},
		{"Null Input", types.StringNull(), 5, false},
		{"Unknown Input", types.StringUnknown(), 5, false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx := context.Background()
			sanitizedValidator := SanitizedNameLengthValidator(test.minLen)

			request := validator.StringRequest{
				ConfigValue: test.input,
			}
			response := &validator.StringResponse{
				Diagnostics: diag.Diagnostics{},
			}

			sanitizedValidator.ValidateString(ctx, request, response)

			if test.expectedErrors {
				assert.True(t, response.Diagnostics.HasError(), "Expected validation to fail")
			} else {
				assert.False(t, response.Diagnostics.HasError(), "Expected validation to succeed")
			}
		})
	}
}

func TestSanitizedLengthValidator_ValidateParameterString(t *testing.T) {
	tests := []struct {
		name           string
		input          types.String
		minLen         int
		expectedError  bool
		expectedErrMsg string
	}{
		{"Valid Input", types.StringValue("valid name"), 5, false, ""},
		{"Invalid Input - Too Short", types.StringValue("abc"), 5, true, "must be at least 5 characters long"},
		{"Valid After Sanitize", types.StringValue("   valid name   "), 5, false, ""},
		{"Invalid After Sanitize", types.StringValue(" a "), 5, true, "must be at least 5 characters long"},
		{"Null Input", types.StringNull(), 5, false, ""},
		{"Unknown Input", types.StringUnknown(), 5, false, ""},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx := context.Background()
			validator := SanitizedNameLengthValidator(test.minLen)

			request := function.StringParameterValidatorRequest{
				Value:            test.input,
				ArgumentPosition: 1,
			}
			response := &function.StringParameterValidatorResponse{}

			validator.ValidateParameterString(ctx, request, response)

			if test.expectedError {
				assert.NotNil(t, response.Error, "Expected validation error")
				assert.Contains(t, response.Error.Error(), test.expectedErrMsg, "Expected error message mismatch")
			} else {
				assert.Nil(t, response.Error, "Expected no validation error")
			}
		})
	}
}

func TestSanitizedLengthValidator_Description(t *testing.T) {
	ctx := context.Background()
	validator := SanitizedNameLengthValidator(5)

	expected := "must be at least 5 characters long"

	assert.Equal(t, expected, validator.Description(ctx))
	assert.Equal(t, expected, validator.MarkdownDescription(ctx))
}

func TestSanitizeInsensitiveModifier_PlanModifyString(t *testing.T) {
	tests := []struct {
		name          string
		stateValue    types.String
		planValue     types.String
		expectedValue types.String
	}{
		{
			name:          "State Null - No Modification on Create",
			stateValue:    types.StringNull(),
			planValue:     types.StringValue("new-name"),
			expectedValue: types.StringValue("new-name"),
		},
		{
			name:          "Plan Null - No Modification on Destroy",
			stateValue:    types.StringValue("old-name"),
			planValue:     types.StringNull(),
			expectedValue: types.StringNull(),
		},
		{
			name:          "Same Value (with extra spaces) After Sanitization",
			stateValue:    types.StringValue("old name"),
			planValue:     types.StringValue("  old   name  "),
			expectedValue: types.StringValue("  old   name  "), // Should match the plan value
		},
		{
			name:          "Different Values After Sanitization - No Override",
			stateValue:    types.StringValue("old name"),
			planValue:     types.StringValue("new name"),
			expectedValue: types.StringValue("new name"),
		},
		{
			name:          "State Null and Plan Unknown",
			stateValue:    types.StringNull(),
			planValue:     types.StringUnknown(),
			expectedValue: types.StringUnknown(), // Plan should stay unchanged
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx := context.Background()
			modifier := SanitizeInsensitiveModifier()

			// Set up request and response
			req := planmodifier.StringRequest{
				StateValue: test.stateValue,
				PlanValue:  test.planValue,
			}
			resp := &planmodifier.StringResponse{
				PlanValue:   test.planValue,
				Diagnostics: diag.Diagnostics{},
			}

			// Perform the modification
			modifier.PlanModifyString(ctx, req, resp)

			// Assert expected value
			assert.Equal(t, test.expectedValue, resp.PlanValue, "Unexpected PlanValue result")

			// Assert no unexpected errors in diagnostics
			assert.False(t, resp.Diagnostics.HasError(), "Expected no errors but found some")
		})
	}
}

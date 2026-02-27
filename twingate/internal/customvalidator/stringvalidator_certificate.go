package customvalidator

import (
	"context"
	"fmt"

	"github.com/Twingate/terraform-provider-twingate/v4/twingate/internal/attr"
	"github.com/Twingate/terraform-provider-twingate/v4/twingate/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ validator.String = certificateValidator{}
)

type certificateValidator struct{}

func (ca certificateValidator) Description(_ context.Context) string {
	return ""
}

func (ca certificateValidator) MarkdownDescription(_ context.Context) string {
	return ""
}

func (ca certificateValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	// Skip on create (state is null) or delete (plan is null).
	if req.Config.Raw.IsNull() {
		return
	}

	var certValue types.String

	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root(attr.Certificate), &certValue)...)

	if resp.Diagnostics.HasError() || certValue.IsNull() || certValue.IsUnknown() {
		return
	}

	_, err := utils.CalculateCertificateFingerprint(certValue.ValueString())
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			path.Root(attr.Certificate),
			"Invalid certificate",
			fmt.Sprintf("Could not calculate fingerprint: %s", err),
		)

		return
	}
}

func Certificate() validator.String {
	return certificateValidator{}
}

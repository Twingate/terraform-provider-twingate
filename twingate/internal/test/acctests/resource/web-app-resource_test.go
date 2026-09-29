package resource

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/Twingate/terraform-provider-twingate/v5/twingate/internal/attr"
	"github.com/Twingate/terraform-provider-twingate/v5/twingate/internal/model"
	"github.com/Twingate/terraform-provider-twingate/v5/twingate/internal/provider/resource"
	"github.com/Twingate/terraform-provider-twingate/v5/twingate/internal/test"
	"github.com/Twingate/terraform-provider-twingate/v5/twingate/internal/test/acctests"
	sdk "github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
)

func webAppResourcePrerequisites(remoteNetworkName, remoteNetworkTFName, x509TFName, certPEM, gatewayTFName, gatewayAddress string) string {
	return fmt.Sprintf(`
	resource "twingate_remote_network" "%s" {
	  name = "%s"
	}
	resource "twingate_x509_certificate_authority" "%s" {
	  name        = "%s"
	  certificate = <<-EOT
%s
	EOT
	}
	resource "twingate_gateway" "%s" {
	  remote_network_id = twingate_remote_network.%s.id
	  address           = "%s"
	  x509_ca_id        = twingate_x509_certificate_authority.%s.id
	}
	`, remoteNetworkTFName, remoteNetworkName, x509TFName, test.RandomName(), certPEM, gatewayTFName, remoteNetworkTFName, gatewayAddress, x509TFName)
}

func terraformResourceWebApp(tfName, gatewayTFName, remoteNetworkTFName, name, address, upstream, downstream, extra string) string {
	return fmt.Sprintf(`
	resource "twingate_web_app_resource" "%s" {
	  name              = "%s"
	  address           = "%s"
	  gateway_id        = twingate_gateway.%s.id
	  remote_network_id = twingate_remote_network.%s.id
	  upstream = {
	    %s
	  }
	  downstream = {
	    %s
	  }
	  %s
	}
	`, tfName, name, address, gatewayTFName, remoteNetworkTFName, upstream, downstream, extra)
}

func webAppStreamBody(port int, tlsMode string) string {
	return fmt.Sprintf("port     = %d\n\t    tls_mode = %q", port, tlsMode)
}

type webAppTestSetup struct {
	prereqs             string
	gatewayTFName       string
	remoteNetworkTFName string
	webAppTFName        string
	theResource         string
	resourceName        string
	resourceAddress     string
}

func newWebAppTestSetup(t *testing.T, gatewayAddress string) webAppTestSetup {
	t.Helper()

	remoteNetworkTFName := test.TerraformRandName("test_rn")
	x509TFName := test.TerraformRandName("test_x509")
	gatewayTFName := test.TerraformRandName("test_gw")
	webAppTFName := test.TerraformRandName("test_web_res")

	return webAppTestSetup{
		prereqs: webAppResourcePrerequisites(test.RandomName(), remoteNetworkTFName, x509TFName,
			acctests.GenerateCACertPEM(t), gatewayTFName, gatewayAddress),
		gatewayTFName:       gatewayTFName,
		remoteNetworkTFName: remoteNetworkTFName,
		webAppTFName:        webAppTFName,
		theResource:         acctests.TerraformWebAppResource(webAppTFName),
		resourceName:        test.RandomName(),
		resourceAddress:     "internal.acme.com",
	}
}

// config leaves `tls_mode` out of both stream blocks, so it also covers the
// attribute falling back to its default.
func (s webAppTestSetup) config(upstreamPort, downstreamPort int, extra string) string {
	return s.configStreams(fmt.Sprintf("port = %d", upstreamPort), fmt.Sprintf("port = %d", downstreamPort), extra)
}

func (s webAppTestSetup) configStreams(upstream, downstream, extra string) string {
	return s.prereqs + terraformResourceWebApp(s.webAppTFName, s.gatewayTFName, s.remoteNetworkTFName,
		s.resourceName, s.resourceAddress, upstream, downstream, extra)
}

func TestAccTwingateWebAppResourceCreate(t *testing.T) {
	t.Parallel()

	setup := newWebAppTestSetup(t, "10.0.0.1:8080")

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		TerraformVersionChecks:   acctests.VersionCheckForWriteOnlyAttributes(),
		CheckDestroy:             acctests.CheckTwingateWebAppResourceDestroy,
		Steps: []sdk.TestStep{
			{
				Config: setup.config(8080, 80, ""),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(setup.theResource),
					sdk.TestCheckResourceAttrSet(setup.theResource, attr.ID),
					sdk.TestCheckResourceAttr(setup.theResource, attr.Name, setup.resourceName),
					sdk.TestCheckResourceAttr(setup.theResource, attr.Address, setup.resourceAddress),
					sdk.TestCheckResourceAttrSet(setup.theResource, attr.GatewayID),
					sdk.TestCheckResourceAttrSet(setup.theResource, attr.RemoteNetworkID),
					sdk.TestCheckResourceAttr(setup.theResource, attr.PathAttr(attr.Upstream, attr.Port), "8080"),
					sdk.TestCheckResourceAttr(setup.theResource, attr.PathAttr(attr.Downstream, attr.Port), "80"),
					sdk.TestCheckResourceAttr(setup.theResource, attr.PathAttr(attr.Upstream, attr.TLSMode), model.TLSClientModeNone),
					sdk.TestCheckResourceAttr(setup.theResource, attr.PathAttr(attr.Downstream, attr.TLSMode), model.TLSServerModeNone),
					sdk.TestCheckNoResourceAttr(setup.theResource, attr.RequestHeaderRewrites),
				),
			},
			{
				ResourceName:      setup.theResource,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// TestAccTwingateWebAppResourceImport imports a resource with every optional
// attribute populated, so `ImportStateVerify` catches anything the import path
// drops.
func TestAccTwingateWebAppResourceImport(t *testing.T) {
	t.Parallel()

	setup := newWebAppTestSetup(t, "10.0.0.8:8080")

	const optionalAttributes = `
	  is_visible = false
	  tags = {
	    env = "prod"
	  }
	  request_header_rewrites = {
	    "X-Twingate-User" = "{{username}}"
	  }`

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		TerraformVersionChecks:   acctests.VersionCheckForWriteOnlyAttributes(),
		CheckDestroy:             acctests.CheckTwingateWebAppResourceDestroy,
		Steps: []sdk.TestStep{
			{
				Config: setup.configStreams(webAppStreamBody(8443, model.TLSClientModeVerifyFull),
					webAppStreamBody(443, model.TLSServerModeTLS13), optionalAttributes),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(setup.theResource),
					sdk.TestCheckResourceAttr(setup.theResource, attr.IsVisible, "false"),
					sdk.TestCheckResourceAttr(setup.theResource, attr.PathAttr(attr.Upstream, attr.TLSMode), model.TLSClientModeVerifyFull),
					sdk.TestCheckResourceAttr(setup.theResource, attr.PathAttr(attr.Downstream, attr.TLSMode), model.TLSServerModeTLS13),
					sdk.TestCheckResourceAttr(setup.theResource, attr.PathAttr(attr.Tags, "env"), "prod"),
					sdk.TestCheckResourceAttr(setup.theResource,
						attr.PathAttr(attr.RequestHeaderRewrites, "X-Twingate-User"), "{{username}}"),
				),
			},
			{
				ResourceName:      setup.theResource,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccTwingateWebAppResourceUpdatePorts(t *testing.T) {
	t.Parallel()

	setup := newWebAppTestSetup(t, "10.0.0.2:8080")

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		TerraformVersionChecks:   acctests.VersionCheckForWriteOnlyAttributes(),
		CheckDestroy:             acctests.CheckTwingateWebAppResourceDestroy,
		Steps: []sdk.TestStep{
			{
				Config: setup.config(8080, 80, ""),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(setup.theResource),
					sdk.TestCheckResourceAttr(setup.theResource, attr.PathAttr(attr.Upstream, attr.Port), "8080"),
				),
			},
			{
				Config: setup.config(9090, 443, ""),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(setup.theResource),
					sdk.TestCheckResourceAttr(setup.theResource, attr.PathAttr(attr.Upstream, attr.Port), "9090"),
					sdk.TestCheckResourceAttr(setup.theResource, attr.PathAttr(attr.Downstream, attr.Port), "443"),
				),
			},
		},
	})
}

func TestAccTwingateWebAppResourceUpdateTLSMode(t *testing.T) {
	t.Parallel()

	setup := newWebAppTestSetup(t, "10.0.0.7:8080")

	emptyPlan := sdk.ConfigPlanChecks{
		PostApplyPostRefresh: []plancheck.PlanCheck{
			plancheck.ExpectEmptyPlan(),
		},
	}

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		TerraformVersionChecks:   acctests.VersionCheckForWriteOnlyAttributes(),
		CheckDestroy:             acctests.CheckTwingateWebAppResourceDestroy,
		Steps: []sdk.TestStep{
			{
				Config:           setup.config(8080, 80, ""),
				ConfigPlanChecks: emptyPlan,
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(setup.theResource),
					sdk.TestCheckResourceAttr(setup.theResource, attr.PathAttr(attr.Upstream, attr.TLSMode), model.TLSClientModeNone),
					sdk.TestCheckResourceAttr(setup.theResource, attr.PathAttr(attr.Downstream, attr.TLSMode), model.TLSServerModeNone),
				),
			},
			{
				Config: setup.configStreams(webAppStreamBody(8443, model.TLSClientModeVerifyFull),
					webAppStreamBody(443, model.TLSServerModeTLS13), ""),
				ConfigPlanChecks: emptyPlan,
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(setup.theResource),
					sdk.TestCheckResourceAttr(setup.theResource, attr.PathAttr(attr.Upstream, attr.TLSMode), model.TLSClientModeVerifyFull),
					sdk.TestCheckResourceAttr(setup.theResource, attr.PathAttr(attr.Downstream, attr.TLSMode), model.TLSServerModeTLS13),
				),
			},
			{
				// Only the upstream changes, so a swapped upstream/downstream
				// mapping shows up as a diff.
				Config: setup.configStreams(webAppStreamBody(8443, model.TLSClientModeInsecure),
					webAppStreamBody(443, model.TLSServerModeTLS13), ""),
				ConfigPlanChecks: emptyPlan,
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(setup.theResource),
					sdk.TestCheckResourceAttr(setup.theResource, attr.PathAttr(attr.Upstream, attr.TLSMode), model.TLSClientModeInsecure),
					sdk.TestCheckResourceAttr(setup.theResource, attr.PathAttr(attr.Downstream, attr.TLSMode), model.TLSServerModeTLS13),
				),
			},
			{
				// Removing `tls_mode` from config must reset both sides to the
				// default rather than keep the value stored by the previous step.
				Config:           setup.config(8080, 80, ""),
				ConfigPlanChecks: emptyPlan,
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(setup.theResource),
					sdk.TestCheckResourceAttr(setup.theResource, attr.PathAttr(attr.Upstream, attr.TLSMode), model.TLSClientModeNone),
					sdk.TestCheckResourceAttr(setup.theResource, attr.PathAttr(attr.Downstream, attr.TLSMode), model.TLSServerModeNone),
				),
			},
		},
	})
}

// TestAccTwingateWebAppResourceHeaderRewrites covers the three shapes the API
// collapses to the same stored value: omitted, explicitly empty, and populated.
// Each step must leave an empty plan, otherwise the attribute drifts forever.
func TestAccTwingateWebAppResourceHeaderRewrites(t *testing.T) {
	t.Parallel()

	setup := newWebAppTestSetup(t, "10.0.0.3:8080")

	// Two entries so the multi-key path is covered. The API stores values as
	// opaque strings, so the second header uses a literal rather than a
	// substitution token.
	const populated = `
	  request_header_rewrites = {
	    "X-Twingate-User"  = "{{username}}"
	    "X-Forwarded-Host" = "internal.acme.com"
	  }`

	emptyPlan := sdk.ConfigPlanChecks{
		PostApplyPostRefresh: []plancheck.PlanCheck{
			plancheck.ExpectEmptyPlan(),
		},
	}

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		TerraformVersionChecks:   acctests.VersionCheckForWriteOnlyAttributes(),
		CheckDestroy:             acctests.CheckTwingateWebAppResourceDestroy,
		Steps: []sdk.TestStep{
			{
				// Omitted: the API stores nothing, state must stay null.
				Config:           setup.config(8080, 80, ""),
				ConfigPlanChecks: emptyPlan,
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(setup.theResource),
					sdk.TestCheckNoResourceAttr(setup.theResource, attr.RequestHeaderRewrites),
				),
			},
			{
				Config:           setup.config(8080, 80, populated),
				ConfigPlanChecks: emptyPlan,
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(setup.theResource),
					sdk.TestCheckResourceAttr(setup.theResource, "request_header_rewrites.%", "2"),
					sdk.TestCheckResourceAttr(setup.theResource,
						attr.PathAttr(attr.RequestHeaderRewrites, "X-Twingate-User"), "{{username}}"),
					sdk.TestCheckResourceAttr(setup.theResource,
						attr.PathAttr(attr.RequestHeaderRewrites, "X-Forwarded-Host"), "internal.acme.com"),
				),
			},
			{
				// Explicitly empty: the API removes the field, but `{}` must be
				// preserved in state rather than collapsing to null.
				Config:           setup.config(8080, 80, "request_header_rewrites = {}"),
				ConfigPlanChecks: emptyPlan,
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(setup.theResource),
					sdk.TestCheckResourceAttr(setup.theResource, "request_header_rewrites.%", "0"),
				),
			},
			{
				// Back to omitted: state must return to null, still without drift.
				Config:           setup.config(8080, 80, ""),
				ConfigPlanChecks: emptyPlan,
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(setup.theResource),
					sdk.TestCheckNoResourceAttr(setup.theResource, attr.RequestHeaderRewrites),
				),
			},
		},
	})
}

// Tags and header rewrites share the same API shape: nothing is stored for an
// empty map, so `tags = {}` must round-trip as an empty map rather than null or
// Terraform rejects the apply as an inconsistent result.
func TestAccTwingateWebAppResourceEmptyTags(t *testing.T) {
	t.Parallel()

	setup := newWebAppTestSetup(t, "10.0.0.21:8080")

	const populated = `
	  tags = {
	    env = "dev"
	  }`

	emptyPlan := sdk.ConfigPlanChecks{
		PostApplyPostRefresh: []plancheck.PlanCheck{
			plancheck.ExpectEmptyPlan(),
		},
	}

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		TerraformVersionChecks:   acctests.VersionCheckForWriteOnlyAttributes(),
		CheckDestroy:             acctests.CheckTwingateWebAppResourceDestroy,
		Steps: []sdk.TestStep{
			{
				// Explicitly empty on create: `{}` must be kept in state.
				Config:           setup.config(8080, 80, "tags = {}"),
				ConfigPlanChecks: emptyPlan,
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(setup.theResource),
					sdk.TestCheckResourceAttr(setup.theResource, attr.Tags+".%", "0"),
				),
			},
			{
				Config:           setup.config(8080, 80, populated),
				ConfigPlanChecks: emptyPlan,
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(setup.theResource),
					sdk.TestCheckResourceAttr(setup.theResource, attr.Tags+".%", "1"),
					sdk.TestCheckResourceAttr(setup.theResource, attr.PathAttr(attr.Tags, "env"), "dev"),
				),
			},
			{
				// Explicitly empty on update: the API clears the tags, state keeps `{}`.
				Config:           setup.config(8080, 80, "tags = {}"),
				ConfigPlanChecks: emptyPlan,
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(setup.theResource),
					sdk.TestCheckResourceAttr(setup.theResource, attr.Tags+".%", "0"),
				),
			},
			{
				// Omitted: state must return to null, still without drift.
				Config:           setup.config(8080, 80, ""),
				ConfigPlanChecks: emptyPlan,
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(setup.theResource),
					sdk.TestCheckNoResourceAttr(setup.theResource, attr.Tags),
				),
			},
		},
	})
}

func TestAccTwingateWebAppResource_InvalidPorts(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name           string
		upstreamPort   int
		downstreamPort int
		expectedErr    *regexp.Regexp
	}{
		{
			name:           "upstream port below range",
			upstreamPort:   0,
			downstreamPort: 80,
			// Keep these patterns short: Terraform hard-wraps diagnostic bodies,
			// so a phrase long enough to span the wrap will never match.
			expectedErr: regexp.MustCompile(`must be between 1 and 65535`),
		},
		{
			name:           "downstream port above range",
			upstreamPort:   8080,
			downstreamPort: 65536,
			expectedErr:    regexp.MustCompile(`must be between 1 and 65535`),
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			setup := newWebAppTestSetup(t, "10.0.0.4:8080")

			sdk.Test(t, sdk.TestCase{
				ProtoV6ProviderFactories: acctests.ProviderFactories,
				PreCheck:                 func() { acctests.PreCheck(t) },
				TerraformVersionChecks:   acctests.VersionCheckForWriteOnlyAttributes(),
				CheckDestroy:             acctests.CheckTwingateWebAppResourceDestroy,
				Steps: []sdk.TestStep{
					{
						Config:      setup.config(c.upstreamPort, c.downstreamPort, ""),
						ExpectError: c.expectedErr,
					},
				},
			})
		})
	}
}

func TestAccTwingateWebAppResource_InvalidTLSMode(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name        string
		upstream    string
		downstream  string
		expectedErr *regexp.Regexp
	}{
		{
			name:        "server mode on upstream",
			upstream:    webAppStreamBody(8080, model.TLSServerModeTLS13),
			downstream:  "port = 80",
			expectedErr: regexp.MustCompile(`value must be one of`),
		},
		{
			name:        "client mode on downstream",
			upstream:    "port = 8080",
			downstream:  webAppStreamBody(443, model.TLSClientModeVerifyFull),
			expectedErr: regexp.MustCompile(`value must be one of`),
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			setup := newWebAppTestSetup(t, "10.0.0.9:8080")

			sdk.Test(t, sdk.TestCase{
				ProtoV6ProviderFactories: acctests.ProviderFactories,
				PreCheck:                 func() { acctests.PreCheck(t) },
				TerraformVersionChecks:   acctests.VersionCheckForWriteOnlyAttributes(),
				CheckDestroy:             acctests.CheckTwingateWebAppResourceDestroy,
				Steps: []sdk.TestStep{
					{
						Config:      setup.configStreams(c.upstream, c.downstream, ""),
						ExpectError: c.expectedErr,
					},
				},
			})
		})
	}
}

func TestAccTwingateWebAppResource_InvalidHeaderRewrites(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name        string
		rewrites    string
		expectedErr *regexp.Regexp
	}{
		{
			name: "keys differing only by case",
			rewrites: `
			  request_header_rewrites = {
			    "X-Twingate-User" = "{{username}}"
			    "x-twingate-user" = "{{username}}"
			  }`,
			expectedErr: regexp.MustCompile(`Duplicate key`),
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			setup := newWebAppTestSetup(t, "10.0.0.5:8080")

			sdk.Test(t, sdk.TestCase{
				ProtoV6ProviderFactories: acctests.ProviderFactories,
				PreCheck:                 func() { acctests.PreCheck(t) },
				TerraformVersionChecks:   acctests.VersionCheckForWriteOnlyAttributes(),
				CheckDestroy:             acctests.CheckTwingateWebAppResourceDestroy,
				Steps: []sdk.TestStep{
					{
						Config:      setup.config(8080, 80, c.rewrites),
						ExpectError: c.expectedErr,
					},
				},
			})
		})
	}
}

func TestAccTwingateWebAppResourceDeleteNonExisting(t *testing.T) {
	t.Parallel()

	setup := newWebAppTestSetup(t, "10.0.0.6:8080")

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		TerraformVersionChecks:   acctests.VersionCheckForWriteOnlyAttributes(),
		CheckDestroy:             acctests.CheckTwingateWebAppResourceDestroy,
		Steps: []sdk.TestStep{
			{
				Config: setup.config(8080, 80, ""),
				Check: acctests.ComposeTestCheckFunc(
					acctests.CheckTwingateResourceExists(setup.theResource),
					acctests.DeleteTwingateResource(setup.theResource, resource.TwingateWebAppResource),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccTwingateWebAppResourceDefaultTags(t *testing.T) {
	t.Parallel()

	setup := newWebAppTestSetup(t, "10.0.3.3:8080")

	userTags := map[string]string{"owner": "example_team", "application": "custom_application"}
	defaultTags := map[string]string{"env": "stage", "application": "default_application"}

	config := setup.config(8080, 80, acctests.TerraformTagsBlock(userTags))

	sdk.Test(t, sdk.TestCase{
		ProtoV6ProviderFactories: acctests.ProviderFactories,
		PreCheck:                 func() { acctests.PreCheck(t) },
		TerraformVersionChecks:   acctests.VersionCheckForWriteOnlyAttributes(),
		CheckDestroy:             acctests.CheckTwingateWebAppResourceDestroy,
		Steps:                    defaultTagsSteps(setup.theResource, config, defaultTags),
	})
}

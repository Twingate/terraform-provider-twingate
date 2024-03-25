// Copyright 2016-2018, Pulumi Corporation.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package twingate

import (
	_ "embed"
	"fmt"
	"path/filepath"

	"github.com/Twingate-Labs/pulumi-twingate/provider/pkg/version"
	"github.com/Twingate/terraform-provider-twingate/twingate"
	pf "github.com/pulumi/pulumi-terraform-bridge/pf/tfbridge"
	"github.com/pulumi/pulumi-terraform-bridge/v3/pkg/tfbridge"
)

//go:embed cmd/pulumi-resource-twingate/bridge-metadata.json
var bridgeMetadata []byte

// all of the token components used below.
const (
	// This variable controls the default name of the package in the package
	// registries for nodejs and python:
	mainPkg = "twingate"
	// modules:
	mainMod = "index" // the root module
)

// Provider returns additional overlaid schema and metadata associated with the provider..
func Provider() tfbridge.ProviderInfo {
	// Instantiate the Terraform provider
	provider := twingate.New(fmt.Sprintf("%s-pulumi", version.Version))()

	// Create a Pulumi provider mapping
	info := tfbridge.ProviderInfo{
		P:                       pf.ShimProvider(provider),
		TFProviderModuleVersion: "6",
		MetadataInfo:            tfbridge.NewProviderMetadata(bridgeMetadata),
		Name:                    "twingate",
		DisplayName:             "Twingate",
		UpstreamRepoPath:        "https://github.com/Twingate/terraform-provider-twingate",
		// The default publisher for all packages is Pulumi.
		// Change this to your personal name (or a company name) that you
		// would like to be shown in the Pulumi Registry if this package is published
		// there.
		Publisher:         "Twingate",
		PluginDownloadURL: "github://api.github.com/twingate-labs",
		// LogoURL is optional but useful to help identify your package in the Pulumi Registry
		// if this package is published there.
		//
		// You may host a logo on a domain you control or add an SVG logo for your package
		// in your repository and use the raw content URL for that file as your logo URL.
		LogoURL:     "",
		Description: "A Pulumi package for creating and managing Twingate cloud resources.",
		// category/cloud tag helps with categorizing the package in the Pulumi Registry.
		// For all available categories, see `Keywords` in
		// https://www.pulumi.com/docs/guides/pulumi-packages/schema/#package.
		Keywords:   []string{"pulumi", "twingate", "category/infrastructure"},
		License:    "Apache-2.0",
		Homepage:   "https://www.twingate.com",
		Repository: "https://github.com/Twingate-Labs/pulumi-twingate",
		Config: map[string]*tfbridge.SchemaInfo{
			"http_timeout": {
				Default: &tfbridge.DefaultInfo{
					Value: 10,
				},
			},
			"http_max_retry": {
				Default: &tfbridge.DefaultInfo{
					Value: 5,
				},
			},
		},
		Resources: map[string]*tfbridge.ResourceInfo{
			"twingate_remote_network":      {Tok: tfbridge.MakeResource(mainPkg, mainMod, "TwingateRemoteNetwork")},
			"twingate_connector":           {Tok: tfbridge.MakeResource(mainPkg, mainMod, "TwingateConnector")},
			"twingate_connector_tokens":    {Tok: tfbridge.MakeResource(mainPkg, mainMod, "TwingateConnectorTokens")},
			"twingate_group":               {Tok: tfbridge.MakeResource(mainPkg, mainMod, "TwingateGroup")},
			"twingate_resource":            {Tok: tfbridge.MakeResource(mainPkg, mainMod, "TwingateResource")},
			"twingate_service_account":     {Tok: tfbridge.MakeResource(mainPkg, mainMod, "TwingateServiceAccount")},
			"twingate_service_account_key": {Tok: tfbridge.MakeResource(mainPkg, mainMod, "TwingateServiceAccountKey")},
			"twingate_user":                {Tok: tfbridge.MakeResource(mainPkg, mainMod, "TwingateUser")},
		},
		DataSources: map[string]*tfbridge.DataSourceInfo{
			"twingate_remote_network":    {Tok: tfbridge.MakeDataSource(mainPkg, mainMod, "getTwingateRemoteNetwork")},
			"twingate_remote_networks":   {Tok: tfbridge.MakeDataSource(mainPkg, mainMod, "getTwingateRemoteNetworks")},
			"twingate_connector":         {Tok: tfbridge.MakeDataSource(mainPkg, mainMod, "getTwingateConnector")},
			"twingate_connectors":        {Tok: tfbridge.MakeDataSource(mainPkg, mainMod, "getTwingateConnectors")},
			"twingate_group":             {Tok: tfbridge.MakeDataSource(mainPkg, mainMod, "getTwingateGroup")},
			"twingate_groups":            {Tok: tfbridge.MakeDataSource(mainPkg, mainMod, "getTwingateGroups")},
			"twingate_resource":          {Tok: tfbridge.MakeDataSource(mainPkg, mainMod, "getTwingateResource")},
			"twingate_resources":         {Tok: tfbridge.MakeDataSource(mainPkg, mainMod, "getTwingateResources")},
			"twingate_user":              {Tok: tfbridge.MakeDataSource(mainPkg, mainMod, "getTwingateUser")},
			"twingate_users":             {Tok: tfbridge.MakeDataSource(mainPkg, mainMod, "getTwingateUsers")},
			"twingate_service_accounts":  {Tok: tfbridge.MakeDataSource(mainPkg, mainMod, "getTwingateServiceAccounts")},
			"twingate_security_policy":   {Tok: tfbridge.MakeDataSource(mainPkg, mainMod, "getTwingateSecurityPolicy")},
			"twingate_security_policies": {Tok: tfbridge.MakeDataSource(mainPkg, mainMod, "getTwingateSecurityPolicies")},
		},
		JavaScript: &tfbridge.JavaScriptInfo{
			PackageName: "@twingate-labs/pulumi-twingate",
			// List any npm dependencies and their versions
			Dependencies: map[string]string{
				"@pulumi/pulumi": "^3.0.0",
			},
			DevDependencies: map[string]string{
				"@types/node": "^10.0.0", // so we can access strongly typed node definitions.
				"@types/mime": "^2.0.0",
			},
		},
		Python: &tfbridge.PythonInfo{
			// List any Python dependencies and their version ranges
			Requires: map[string]string{
				"pulumi": ">=3.0.0,<4.0.0",
			},
		},
		Golang: &tfbridge.GolangInfo{
			ImportBasePath: filepath.Join(
				fmt.Sprintf("github.com/Twingate-Labs/pulumi-%[1]s/sdk/", mainPkg),
				tfbridge.GetModuleMajorVersion(version.Version),
				"go",
				mainPkg,
			),
			GenerateResourceContainerTypes: true,
		},
		CSharp: &tfbridge.CSharpInfo{
			PackageReferences: map[string]string{
				"Pulumi": "3.*",
			},
			RootNamespace: "TwingateLabs",
		},
		GitHubOrg: "Twingate",
	}

	info.SetAutonaming(255, "-")

	return info
}

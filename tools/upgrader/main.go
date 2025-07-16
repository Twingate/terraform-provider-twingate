package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
)

const (
	terraformFileExtension = ".tf"

	versionFlag = "--version"

	resourceType     = "resource"
	twingateResource = "twingate_resource"

	protocolsBlock = "protocols"
	tcpBlock       = "tcp"
	udpBlock       = "udp"

	v2AccessBlock = "access"
	v2GroupIDs    = "group_ids"
	v2ServiceIDs  = "service_account_ids"

	v3AccessGroupBlock   = "access_group"
	v3AccessServiceBlock = "access_service"
	v3GroupID            = "group_id"
	v3ServiceID          = "service_account_id"

	migrationNotRequired = 0
	migrationV1          = 1
	migrationV2          = 2
)

var (
	commit  = "dev"
	version = "dev"
)

var terraformFiles []string

func main() {
	if len(os.Args) < 2 {
		fmt.Printf("Usage: %s <path to terraform file or folder>\n", os.Args[0])
		os.Exit(1)
	}

	path := os.Args[1]

	if path == versionFlag {
		fmt.Printf("Twingate Upgrader v%s (commit: %s)\n", version, commit)
		os.Exit(0)
	}

	info, err := os.Stat(path)
	if err != nil {
		fmt.Printf("Failed to get file info: %v\n", err)
		os.Exit(1)
	}

	if info.IsDir() {
		err := filepath.WalkDir(path, collectTerraformFiles)
		if err != nil {
			fmt.Printf("Error walking directory: %v\n", err)
			os.Exit(1)
		}
	} else if isTerraformFile(path) {
		terraformFiles = append(terraformFiles, path)
	}

	if len(terraformFiles) == 0 {
		fmt.Printf("Not recognized file type. Please provide a path to a terraform file or folder.\n")
		os.Exit(1)
	}

	minRequiredMigration := migrationNotRequired

	for _, file := range terraformFiles {
		version, required := requiredMigration(file)
		if required && (version < minRequiredMigration || minRequiredMigration == migrationNotRequired) {
			minRequiredMigration = version
		}
	}

	if minRequiredMigration == migrationNotRequired {
		fmt.Println("No migration required.")
		return
	}

	hasModified := false
	for _, file := range terraformFiles {
		hasModified = applyMigration(file, minRequiredMigration) || hasModified
	}

	if hasModified {
		fmt.Println("------------------------------------------------------------------------------")
		fmt.Printf("Modified files were upgraded to version v%d, now you need to run: $ terraform apply \n", minRequiredMigration+1)
		fmt.Println("and then run upgrader tool again to check if there are any other changes that need to be applied.")
		fmt.Println("------------------------------------------------------------------------------")
	}
}

func applyMigration(file string, migration int) (changesSaved bool) {
	input := readFile(file)
	f := parseFile(input)

	migrationMap := map[int]func(body *hclwrite.Body) (hasChanges bool){
		migrationV1: applyMigrationV1,
		migrationV2: applyMigrationV2,
	}

	migrationFunc := migrationMap[migration]

	if migrationFunc == nil {
		fmt.Printf("Migration function not found for migration v%d", migration)
		os.Exit(1)
	}

	if runBlocksProcessor(f.Body(), migrationFunc) {
		result := string(f.Bytes())

		fmt.Printf("\n--------[Please check changes for the file (applying migration from v%d to v%d and some formatting): %s]-----------\n", migration, migration+1, file)
		fmt.Println(getUnifiedDiff(string(input), result))
		fmt.Println("------------------------------------------------------------------------------")

		if strings.ToLower(getUserResponse("Do you want to save the changes? (y/n): ")) == "y" {
			saveResults(file, result)
			fmt.Println("Changes saved successfully!")

			return true
		} else {
			fmt.Println("Changes were not saved.")
		}
	}

	return false
}

func requiredMigration(file string) (int, bool) {
	content := readFile(file)
	f := parseFile(content)

	if runBlocksProcessor(f.Body(), requiresMigrationV1) {
		return migrationV1, true
	}

	if runBlocksProcessor(f.Body(), requiresMigrationV2) {
		return migrationV2, true
	}

	return migrationNotRequired, false
}

func readFile(file string) []byte {
	content, err := os.ReadFile(file)
	if err != nil {
		fmt.Printf("Failed to read file %s: %v", file, err)
		os.Exit(1)
	}

	return content
}

func parseFile(content []byte) *hclwrite.File {
	f, diags := hclwrite.ParseConfig(content, "input.tf", hcl.Pos{Line: 1, Column: 1})
	if diags.HasErrors() {
		for _, diag := range diags {
			if diag.Subject != nil {
				fmt.Printf("[%s:%d] %s: %s\n", diag.Subject.Filename, diag.Subject.Start.Line, diag.Summary, diag.Detail)
			} else {
				fmt.Printf("%s: %s\n", diag.Summary, diag.Detail)
			}
		}

		os.Exit(1)
	}

	if f == nil {
		fmt.Println("[ERROR]: failed to parse terraform file.")
		os.Exit(1)
	}

	return f
}

func isTerraformFile(path string) bool {
	return strings.HasSuffix(path, terraformFileExtension)
}

func collectTerraformFiles(path string, d os.DirEntry, err error) error {
	if err != nil {
		return err
	}

	if !d.IsDir() && isTerraformFile(path) {
		terraformFiles = append(terraformFiles, path)
	}

	return nil
}

func getUserResponse(question string) string {
	fmt.Printf("\n> %s", question)

	var resp string
	_, err := fmt.Scanln(&resp)
	if err != nil {
		panic(fmt.Errorf("Failed to read user response: %w", err))
	}

	return resp
}

func saveResults(filePath string, result string) {
	err := os.WriteFile(filePath, []byte(result), 0644)
	if err != nil {
		fmt.Printf("Error saving file: %v\n", err)
		os.Exit(1)
	}
}

func runBlocksProcessor(body *hclwrite.Body, processor func(body *hclwrite.Body) (hasChanges bool)) (hasChanges bool) {
	blocks := body.Blocks()
	for _, block := range blocks {
		if block.Type() == resourceType && block.Labels()[0] == twingateResource {
			hasChanges = processor(block.Body()) || hasChanges
		}
	}

	return hasChanges
}

func requiresMigrationV1(body *hclwrite.Body) bool {
	for _, block := range body.Blocks() {
		if block.Type() == protocolsBlock {
			return true
		}
	}

	return false
}

func applyMigrationV1(body *hclwrite.Body) (hasChanges bool) {
	attributes := map[string]hclwrite.Tokens{}
	var (
		protocols *hclwrite.Block
		tcp       *hclwrite.Block
		udp       *hclwrite.Block
	)

	for _, block := range body.Blocks() {
		if block.Type() == protocolsBlock {
			protocols = block

			for name, attr := range block.Body().Attributes() {
				attributes[name] = attr.Expr().BuildTokens(nil)
			}

			for _, innerBlock := range block.Body().Blocks() {
				switch innerBlock.Type() {
				case tcpBlock:
					tcp = innerBlock
				case udpBlock:
					udp = innerBlock
				}
			}

		}
	}

	if protocols == nil {
		return false
	}

	body.RemoveBlock(protocols)

	newBlock := hclwrite.NewBlock(protocolsBlock, nil)

	if tcp != nil {
		newBlock.Body().SetAttributeRaw(tcpBlock, tcp.BuildTokens(nil)[1:])
	}

	if udp != nil {
		newBlock.Body().SetAttributeRaw(udpBlock, udp.BuildTokens(nil)[1:])
	}

	for name, attr := range attributes {
		newBlock.Body().SetAttributeRaw(name, attr)
	}

	body.SetAttributeRaw(protocolsBlock, newBlock.BuildTokens(nil)[1:])

	return true
}

func requiresMigrationV2(body *hclwrite.Body) bool {
	for _, block := range body.Blocks() {
		if block.Type() == v2AccessBlock {
			return true
		}
	}

	return false
}

func applyMigrationV2(body *hclwrite.Body) (hasChanges bool) {
	migrationMap := map[string]struct {
		blockName string
		attribute string
	}{
		v2GroupIDs: {
			blockName: v3AccessGroupBlock,
			attribute: v3GroupID,
		},
		v2ServiceIDs: {
			blockName: v3AccessServiceBlock,
			attribute: v3ServiceID,
		},
	}

	accessAttributes := map[string]hclwrite.Tokens{}
	var accessBlock *hclwrite.Block

	for _, block := range body.Blocks() {
		if block.Type() == v2AccessBlock {
			accessBlock = block

			for name, attr := range block.Body().Attributes() {
				accessAttributes[name] = attr.Expr().BuildTokens(nil)
			}
		}
	}

	if accessBlock == nil {
		return false
	}

	body.RemoveBlock(accessBlock)

	for name, tokens := range accessAttributes {
		if len(tokens) == 0 {
			continue
		}

		migration := migrationMap[name]

		// Add the `dynamic "access_group"` block
		dynamicBlock := body.AppendNewBlock("dynamic", []string{migration.blockName})
		dynamicBlock.Body().SetAttributeRaw("for_each", tokens)

		// Add the inner `content` block
		contentBlock := dynamicBlock.Body().AppendNewBlock("content", nil)

		// Set attributes inside the `content` block
		contentBlock.Body().SetAttributeRaw(migration.attribute,
			hclwrite.TokensForIdentifier(fmt.Sprintf("%s.value", migration.blockName)))

		// Example:
		// dynamic "access_group" {
		//    for_each = [twingate_group.infra.id, twingate_group.security.id]
		//    content {
		//      group_id = access_group.value
		//    }
		//  }
	}

	return true
}

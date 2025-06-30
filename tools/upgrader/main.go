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
)

var terraformFiles []string

func main() {
	if len(os.Args) < 2 {
		fmt.Printf("Usage: %s <path to terraform file or folder>\n", os.Args[0])
		os.Exit(1)
	}

	path := os.Args[1]

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

	for _, file := range terraformFiles {
		// todo: check minimum migration version required

		version := requiredMigration(file)
	}

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

func walkDirFn(path string, d os.DirEntry, err error) error {
	if err != nil {
		return err
	}

	if !d.IsDir() && isTerraformFile(path) {
		processFile(path)
	}

	return nil
}

func processFile(filePath string) {
	input, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Printf("Failed to read file %s: %v", filePath, err)
		os.Exit(1)
	}

	result, hasChanges := modifyFile(input)
	if result == "" || !hasChanges {
		return
	}

	fmt.Printf("\n----------------[Please check changes for the file: %s]-----------------------\n", filePath)
	fmt.Println(getUnifiedDiff(string(input), result))
	fmt.Println("------------------------------------------------------------------------------")

	if strings.ToLower(getUserResponse("Do you want to save the changes? (y/n): ")) == "y" {
		saveResults(filePath, result)
		fmt.Println("Changes saved successfully!")
	} else {
		fmt.Println("Changes were not saved.")
	}
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

func modifyFile(src []byte) (result string, hasChanges bool) {
	f, diags := hclwrite.ParseConfig(src, "input.tf", hcl.Pos{Line: 1, Column: 1})
	if diags.HasErrors() {
		for _, diag := range diags {
			if diag.Subject != nil {
				fmt.Printf("[%s:%d] %s: %s\n", diag.Subject.Filename, diag.Subject.Start.Line, diag.Summary, diag.Detail)
			} else {
				fmt.Printf("%s: %s\n", diag.Summary, diag.Detail)
			}
		}

		return "", false
	}

	if f == nil {
		fmt.Println("[ERROR]: failed to parse terraform file.")
		return "", false
	}

	hasChanges = processResourceBlocks(f.Body())

	return string(f.Bytes()), hasChanges
}

func processResourceBlocks(body *hclwrite.Body) (hasChangesV1, hasChangesV2 bool) {
	blocks := body.Blocks()
	for _, block := range blocks {
		if block.Type() == resourceType && block.Labels()[0] == twingateResource {
			hasChangesV1 = applyMigrationV1(block.Body()) || hasChangesV1

			hasChangesV1 = applyMigrationV2(block.Body()) || hasChangesV1
		}
	}

	return hasChangesV1
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

	if protocols != nil {
		body.RemoveBlock(protocols)

		hasChanges = true
	}

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

	return hasChanges
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

	if accessBlock != nil {
		body.RemoveBlock(accessBlock)

		hasChanges = true
	}

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

	return hasChanges
}

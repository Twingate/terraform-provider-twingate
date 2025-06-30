package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
)

var (
	commit  = "dev"
	version = "dev"
)

const (
	terraformFileExtension = ".tf"

	versionFlag = "--version"
)

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
		err := filepath.WalkDir(path, walkDirFn)
		if err != nil {
			fmt.Printf("Error walking directory: %v\n", err)
			os.Exit(1)
		}

		return
	}

	if isTerraformFile(path) {
		processFile(path)
		return
	}

	fmt.Printf("Not recognized file type. Please provide a path to a terraform file or folder.\n")
	os.Exit(1)
}

func isTerraformFile(path string) bool {
	return strings.HasSuffix(path, terraformFileExtension)
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

func processAttributes(body *hclwrite.Body) (hasChanges bool) {
	accessTokens := map[string]hclwrite.Tokens{}
	var accessBlock *hclwrite.Block

	for i, block := range body.Blocks() {
		if block.Type() == "access" {
			accessBlock = body.Blocks()[i]
			for name, attr := range block.Body().Attributes() {
				list := strings.TrimSpace(string(attr.Expr().BuildTokens(nil).Bytes()))

				values := []string{}
				for _, v := range strings.Split(list[1:len(list)-1], ",") {
					values = append(values, strings.TrimSpace(v))
				}

				accessTokens[name] = attr.Expr().BuildTokens(nil)
			}
		}
	}

	if accessBlock != nil {
		body.RemoveBlock(accessBlock)

		hasChanges = true
	}

	if len(accessTokens["group_ids"]) > 0 {
		// Add the `dynamic "access_group"` block
		dynamicBlock := body.AppendNewBlock("dynamic", []string{"access_group"})
		dynamicBlock.Body().SetAttributeRaw("for_each", accessTokens["group_ids"])

		// Add the inner `content` block
		contentBlock := dynamicBlock.Body().AppendNewBlock("content", nil)

		// Set attributes inside the `content` block
		contentBlock.Body().SetAttributeRaw("group_id", hclwrite.TokensForIdentifier("access_group.value"))

		// Example:
		// dynamic "access_group" {
		//    for_each = [twingate_group.infra.id, twingate_group.security.id]
		//    content {
		//      group_id = access_group.value
		//    }
		//  }

		hasChanges = true
	}

	if len(accessTokens["service_account_ids"]) > 0 {
		// Add the `dynamic "access_service"` block
		dynamicBlock := body.AppendNewBlock("dynamic", []string{"access_service"})
		dynamicBlock.Body().SetAttributeRaw("for_each", accessTokens["service_account_ids"])

		// Add the inner `content` block
		contentBlock := dynamicBlock.Body().AppendNewBlock("content", nil)

		// Set attributes inside the `content` block
		contentBlock.Body().SetAttributeRaw("service_account_id", hclwrite.TokensForIdentifier("access_service.value"))

		// Example:
		// dynamic "access_service" {
		//    for_each = [twingate_service_account.infra.id, twingate_service_account.security.id]
		//    content {
		//      service_account_id = access_service.value
		//    }
		//  }

		hasChanges = true
	}

	return hasChanges
}

func processResourceBlocks(body *hclwrite.Body) (hasChanges bool) {
	blocks := body.Blocks()
	for _, block := range blocks {
		if block.Type() == "resource" && block.Labels()[0] == "twingate_resource" {
			hasChanges = processAttributes(block.Body()) || hasChanges
		}
	}

	return hasChanges
}

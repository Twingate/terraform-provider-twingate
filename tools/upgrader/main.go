package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Printf("Usage: %s <path to terraform file>\n", os.Args[0])
		os.Exit(1)
	}

	filePath := os.Args[1]
	input, err := os.ReadFile(filePath)
	if err != nil {
		panic(fmt.Sprintf("Failed to read file %s: %v", filePath, err))
	}

	result := processFile(input)
	if result == "" {
		return
	}

	fmt.Println("----------------------------------------")
	fmt.Println(getUnifiedDiff(string(input), result))
	fmt.Println("----------------------------------------")

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

func processFile(src []byte) string {
	f, diags := hclwrite.ParseConfig(src, "input.tf", hcl.Pos{Line: 1, Column: 1})
	if diags.HasErrors() {
		for _, diag := range diags {
			if diag.Subject != nil {
				fmt.Printf("[%s:%d] %s: %s\n", diag.Subject.Filename, diag.Subject.Start.Line, diag.Summary, diag.Detail)
			} else {
				fmt.Printf("%s: %s\n", diag.Summary, diag.Detail)
			}
		}

		return ""
	}

	if f == nil {
		fmt.Println("Failed to parse HCL.")
		return ""
	}

	processResourceBlocks(f.Body())

	return string(f.Bytes())
}

func processAttributes(body *hclwrite.Body) {
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
	}
}

func processResourceBlocks(body *hclwrite.Body) {
	blocks := body.Blocks()
	for _, block := range blocks {
		if block.Type() == "resource" && block.Labels()[0] == "twingate_resource" {
			processAttributes(block.Body())
		}
	}
}

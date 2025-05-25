package main

import (
	"fmt"
	"os"
	"strings"
)

func main() {
	getAPIKey()
	if len(os.Args) < 2 {
		fmt.Printf("Usage: %s <path to terraform file>\n", os.Args[0])
		os.Exit(1)
	}

	filePath := os.Args[1]
	input, err := os.ReadFile(filePath)
	if err != nil {
		panic(fmt.Sprintf("Failed to read file %s: %v", filePath, err))
	}

	result := callLLM(string(input))

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

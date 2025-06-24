package main

import (
	"fmt"
	"github.com/pmezard/go-difflib/difflib"
)

// splitLines splits a string into lines
func splitLines(data string) []string {
	return difflib.SplitLines(data)
}

func getUnifiedDiff(original, modified string) string {
	// Convert file content into slices of lines
	originalLines := splitLines(original)
	modifiedLines := splitLines(modified)

	// Create the unified diff
	unifiedDiff := difflib.UnifiedDiff{
		A:        originalLines,
		B:        modifiedLines,
		FromFile: "original",
		ToFile:   "modified",
		Context:  3, // Number of contextual lines before/after changes
	}

	// Generate the diff string
	diff, err := difflib.GetUnifiedDiffString(unifiedDiff)
	if err != nil {
		panic(fmt.Errorf("Error creating unified diff: %w", err))
	}

	return diff
}

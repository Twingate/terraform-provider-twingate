package attr

import "strings"

const (
	attrFirstElement  = ".0"
	attrPathSeparator = ".0."
	attrLenSymbol     = ".#"
)

func First(attributes ...string) string {
	attr := Path(attributes...)

	if attr == "" {
		return ""
	}

	return attr + attrFirstElement
}

func Path(attributes ...string) string {
	return strings.Join(attributes, attrPathSeparator)
}

func Len(attributes ...string) string {
	attr := Path(attributes...)

	if attr == "" {
		return ""
	}

	return attr + attrLenSymbol
}

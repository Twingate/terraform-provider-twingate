package attr

import "strings"

const (
	attrPathSeparator = ".0."
	attrLenSymbol     = ".#"
)

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

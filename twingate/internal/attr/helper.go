package attr

import "strings"

const (
	attrFirstElement  = ".0"
	attrPathSeparator = ".0."
	attrSeparator     = "."
	attrLenSymbol     = ".#"
)

func First(attributes ...string) string {
	attr := Path(attributes...)

	if attr == "" {
		return ""
	}

	return attr + attrFirstElement
}

func FirstAttr(attributes ...string) string {
	attr := PathAttr(attributes...)

	if attr == "" {
		return ""
	}

	return attr + attrFirstElement
}

func Path(attributes ...string) string {
	return strings.Join(attributes, attrPathSeparator)
}

func PathAttr(attributes ...string) string {
	return strings.Join(attributes, attrSeparator)
}

func Len(attributes ...string) string {
	attr := Path(attributes...)

	if attr == "" {
		return ""
	}

	return attr + attrLenSymbol
}

func LenAttr(attributes ...string) string {
	attr := PathAttr(attributes...)

	if attr == "" {
		return ""
	}

	return attr + attrLenSymbol
}

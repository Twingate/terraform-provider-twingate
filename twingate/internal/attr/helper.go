package attr

import "strings"

const (
	attrFirstElement       = ".0"
	attrPathSeparator      = ".0."
	attrLenSymbol          = ".#"
	attrBlockPathSeparator = "."
)

func First(attributes ...string) string {
	return attrFirst(Path, attributes...)
}

func BlockFirst(attributes ...string) string {
	return attrFirst(BlockPath, attributes...)
}

func Path(attributes ...string) string {
	return strings.Join(attributes, attrPathSeparator)
}

func BlockPath(attributes ...string) string {
	return strings.Join(attributes, attrBlockPathSeparator)
}

func Len(attributes ...string) string {
	return attrLen(Path, attributes...)
}

func BlockLen(attributes ...string) string {
	return attrLen(BlockPath, attributes...)
}

type pathFuncType func(attributes ...string) string

func attrLen(pathFunc pathFuncType, attributes ...string) string {
	attr := pathFunc(attributes...)

	if attr == "" {
		return ""
	}

	return attr + attrLenSymbol
}

func attrFirst(pathFunc pathFuncType, attributes ...string) string {
	attr := pathFunc(attributes...)

	if attr == "" {
		return ""
	}

	return attr + attrFirstElement
}

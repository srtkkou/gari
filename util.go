package gari

import (
	"regexp"
	"strings"
)

var (
	// アンダーバーの正規表現
	reUnderbar = regexp.MustCompile(`_(.)`)
)

// スネークケースを小文字キャメルケースに変換する。
// * snake_case -> lowerCamelCase
func snakeToLowerCamelCase(str string) string {
	return strings.ReplaceAll(
		reUnderbar.ReplaceAllStringFunc(str, strings.ToUpper),
		"_", "")
}

// スネークケースを大文字キャメルケースに変換する。
// * snake_case -> UpperCamelCase
func snakeToUpperCamelCase(str string) string {
	s := snakeToLowerCamelCase(str)
	return strings.ToUpper(s[0:1]) + s[1:]
}

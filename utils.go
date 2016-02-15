package shell

import "regexp"

var (
	escape = regexp.MustCompile(`[^\w/]`)
)

func Escape(argument string) string {
	return escape.ReplaceAllString(argument, `\$0`)
}

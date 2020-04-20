package tool

import (
	"strings"
)

func Replace(temp string, args ...string) (result string) {
	const placeholder = "{}"
	result = temp
	for _, arg := range args {
		result = strings.Replace(result, placeholder, arg, 1)
	}
	return
}

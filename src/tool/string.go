package tool

import (
	"strconv"
	"strings"
)

func Replace(temp string, args ...string) string {
	for n, arg := range args {
		placeholder := "{" + strconv.Itoa(n) + "}"
		temp = strings.ReplaceAll(temp, placeholder, arg)
	}
	return temp
}

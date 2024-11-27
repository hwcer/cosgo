package registry

import "strings"

var Formatter func(string) string = strings.ToLower

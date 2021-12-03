package controllers

import (
	"fmt"
	"strings"
)

func boolToString(value *bool) (result string) {
	switch *value {
	case true:
		return "1"
	case false:
		return "0"
	default:
		return "default"
	}
}

func checkArgumentsToString(args []string) string {
	if args == nil || len(args) == 0 {
		return ""
	}

	return fmt.Sprintf("!%s", strings.Join(args, "!"))
}

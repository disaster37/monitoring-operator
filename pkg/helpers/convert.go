package helpers

import (
	"fmt"
	"strings"
)

func BoolToString(value *bool) (result string) {

	if value == nil {
		return "2"
	}

	switch *value {
	case true:
		return "1"
	case false:
		return "0"
	default:
		return "2"
	}
}

func CheckArgumentsToString(args []string) string {
	if args == nil || len(args) == 0 {
		return ""
	}

	return fmt.Sprintf("!%s", strings.Join(args, "!"))
}

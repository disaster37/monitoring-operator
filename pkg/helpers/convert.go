package helpers

import (
	"fmt"
	"strings"
)

func BoolToString(value *bool) (result string) {
	if value == nil {
		return "2"
	}

	if *value {
		return "1"
	}
	return "0"
}

func StringToBool(value string) (result *bool) {
	trueValue := true
	falseValue := false
	if value == "1" {
		return &trueValue
	} else if value == "2" {
		return nil
	} else {
		return &falseValue
	}
}

func StringToSlice(value, separator string) (result []string) {
	if value == "" {
		return []string{}
	}
	result = strings.Split(value, separator)
	for i, s := range result {
		result[i] = strings.TrimSpace(s)
	}
	return result
}

func CheckArgumentsToString(args []string) string {
	if len(args) == 0 {
		return ""
	}

	return fmt.Sprintf("!%s", strings.Join(args, "!"))
}

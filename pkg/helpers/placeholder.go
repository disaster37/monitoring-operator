package helpers

import (
	"fmt"
	"strings"
)

func PlaceholdersInString(str string, values map[string]string) (result string) {
	if str == "" || values == nil || len(values) == 0 {
		return str
	}

	for key, value := range values {
		str = strings.ReplaceAll(str, fmt.Sprintf("<%s>", key), value)
	}

	return str
}

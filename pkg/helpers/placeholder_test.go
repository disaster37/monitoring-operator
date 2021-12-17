package helpers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPlaceholdersInString(t *testing.T) {
	var (
		str    string
		values map[string]string
	)

	// When empty
	str = PlaceholdersInString(str, values)
	assert.Equal(t, "", str)

	str = "plop"
	values = map[string]string{}
	str = PlaceholdersInString(str, values)
	assert.Equal(t, "plop", str)

	// When no placeholders match
	str = "plop"
	values = map[string]string{
		"name":      "test-name",
		"namespace": "test-namespace",
	}
	str = PlaceholdersInString(str, values)
	assert.Equal(t, "plop", str)

	// When placeholders match
	str = "plop <namespace>/<name> on <namespace>"
	values = map[string]string{
		"name":      "test-name",
		"namespace": "test-namespace",
	}
	str = PlaceholdersInString(str, values)
	assert.Equal(t, "plop test-namespace/test-name on test-namespace", str)
}

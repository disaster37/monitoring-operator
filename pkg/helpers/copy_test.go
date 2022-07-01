package helpers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCopyMapString(t *testing.T) {
	current := map[string]string{
		"aa": "bb",
		"dd": "ee",
	}
	expected := map[string]string{
		"aa": "bb",
		"dd": "ee",
	}

	assert.Equal(t, expected, CopyMapString(current))
	assert.Nil(t, CopyMapString(nil))
}

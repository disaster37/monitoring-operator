package helpers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBoolToString(t *testing.T) {
	trueValue := true
	falseValue := false

	assert.Equal(t, "1", BoolToString(&trueValue))
	assert.Equal(t, "0", BoolToString(&falseValue))
	assert.Equal(t, "2", BoolToString(nil))
}

func TestStringToBool(t *testing.T) {

	assert.True(t, *StringToBool("1"))
	assert.False(t, *StringToBool("0"))
	assert.False(t, *StringToBool("fake"))
	assert.Nil(t, StringToBool("2"))
}

func TestCheckArgumentsToString(t *testing.T) {

	assert.Equal(t, "", CheckArgumentsToString(nil))
	assert.Equal(t, "", CheckArgumentsToString([]string{}))
	assert.Equal(t, "!arg1", CheckArgumentsToString([]string{"arg1"}))
	assert.Equal(t, "!arg1!arg2", CheckArgumentsToString([]string{"arg1", "arg2"}))
}

func TestStringToSlice(t *testing.T) {

	assert.Equal(t, []string{"test"}, StringToSlice("test", ","))
	assert.Equal(t, []string{}, StringToSlice("", ","))
	assert.Equal(t, []string{"test", "test2"}, StringToSlice("test,test2", ","))
	assert.Equal(t, []string{"test", "test2"}, StringToSlice("test, test2", ","))

}

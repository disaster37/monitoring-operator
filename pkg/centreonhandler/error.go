package centreonhandler

import "regexp"

func IsErrorNotFound(providedErr error) bool {

	r, err := regexp.Compile(`Object not found`)
	if err != nil {
		panic(err)
	}

	if r.MatchString(providedErr.Error()) {
		return true
	}
	return false
}

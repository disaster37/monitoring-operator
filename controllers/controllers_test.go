package controllers

import "github.com/stretchr/testify/assert"

func (t *ControllerTestSuite) TestIsRouteCRD() {
	ok, err := IsRouteCRD(t.cfg)
	assert.NoError(t.T(), err)
	assert.True(t.T(), ok)
}

package handler

import (
	"testing"

	"github.com/izziiyt/compaa/component"
	"gotest.tools/v3/assert"
)

func Test_RequirementsTXTLookUp(t *testing.T) {
	h := &RequirementsTXT{}
	as, err := h.LookUp("testdata/requirements.txt")
	assert.NilError(t, err)
	assert.Equal(t, len(as), 3)

	m0 := as[0].(*component.Module)
	assert.Equal(t, m0.Name, "requests")
	m1 := as[1].(*component.Module)
	assert.Equal(t, m1.Name, "PyYAML")
	m2 := as[2].(*component.Module)
	assert.Equal(t, m2.Name, "pytz")
}

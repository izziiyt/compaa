package handler

import (
	"testing"

	"github.com/izziiyt/compaa/component"
	"gotest.tools/v3/assert"
)

func Test_PackageJSONLookUp(t *testing.T) {
	h := &PackageJSON{}
	as, err := h.LookUp("testdata/package.json")
	assert.NilError(t, err)
	assert.Equal(t, len(as), 3)

	m0 := as[0].(*component.Module)
	assert.Equal(t, m0.Name, "abc")
	m1 := as[1].(*component.Module)
	assert.Equal(t, m1.Name, "aws-sdk")
	m2 := as[2].(*component.Module)
	assert.Equal(t, m2.Name, "minimist")
}

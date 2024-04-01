package handler

import (
	"testing"

	"github.com/izziiyt/compaa/component"
	"gotest.tools/v3/assert"
)

func Test_GemfileLookUp(t *testing.T) {
	h := GemFile{}
	as, err := h.LookUp("testdata/Gemfile")
	assert.NilError(t, err)
	assert.Equal(t, len(as), 7)
	l := as[0].(*component.Language)
	assert.Equal(t, l.Name, "ruby")
	assert.Equal(t, l.Version, "3.2.2")
	m := as[1].(*component.Module)
	assert.Equal(t, m.Name, "rails")
	m = as[len(as)-1].(*component.Module)
	assert.Equal(t, m.Name, "spring")
}

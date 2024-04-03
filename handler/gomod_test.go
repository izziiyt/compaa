package handler

import (
	"testing"

	"github.com/izziiyt/compaa/component"
	"gotest.tools/v3/assert"
)

func Test_GoModLookUp(t *testing.T) {
	h := &GoMod{}
	as, err := h.LookUp("testdata/gomod")
	assert.NilError(t, err)
	assert.Equal(t, len(as), 7)

	l := as[0].(*component.Language)
	assert.Equal(t, l.Name, "go")

	m0 := as[1].(*component.Module)
	assert.Equal(t, m0.Name, "github.com/sample/example")
	m1 := as[2].(*component.Module)
	assert.Equal(t, m1.Name, "go.uber.org/zap")
	m2 := as[3].(*component.Module)
	assert.Equal(t, m2.Name, "golang.org/x/text")
	m3 := as[4].(*component.Module)
	assert.Equal(t, m3.Name, "gopkg.in/go-playground/validator.v8")
}

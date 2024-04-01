package handler

import (
	"testing"

	"github.com/izziiyt/compaa/component"
	"gotest.tools/v3/assert"
)

func Test_DockerfileLookUp(t *testing.T) {
	h := &Dockerfile{}
	as, err := h.LookUp("testdata/Dockerfile")
	assert.NilError(t, err)
	assert.Equal(t, len(as), 2)

	i0 := as[0].(*component.Image)
	assert.Equal(t, i0.Registry, "docker.io")
	assert.Equal(t, i0.Namespace, "library")
	assert.Equal(t, i0.Repository, "golang")

	i1 := as[1].(*component.Image)
	assert.Equal(t, i1.Registry, "gcr.io")
	assert.Equal(t, i1.Namespace, "distroless")
	assert.Equal(t, i1.Repository, "base-nossl-debian11")
}

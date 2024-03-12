package handler

import (
	"context"
	"github.com/izziiyt/compaa/component"
)

type Handler interface {
	Handle(ctx context.Context, path string)
	LookUp(ctx context.Context, path string) ([]component.Component, error)
}

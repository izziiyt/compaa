package handler

import (
	"compaa/component"
	"context"
)

type Handler interface {
	Handle(ctx context.Context, path string)
	LookUp(ctx context.Context, path string) ([]component.Component, error)
}

package handler

import (
	"context"
	"depeol/component"
)

type Handler interface {
	Handle(ctx context.Context, path string)
	LookUp(ctx context.Context, path string) ([]component.Component, error)
}

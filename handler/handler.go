package handler

import (
	"context"
	"fmt"
	"sync"

	"github.com/fatih/color"
	"github.com/izziiyt/compaa/component"
)

type Handler interface {
	LookUp(path string) ([]component.Component, error)
	SyncWithSource(c component.Component, ctx context.Context) component.Component
}

func Handle(h Handler, ctx context.Context, path string, wc *component.WarnCondition) {
	fmt.Printf("%v\n", path)

	cs, err := h.LookUp(path)
	if err != nil {
		color.Red("├ LookUp error: %v\n", err)
	}

	wg := &sync.WaitGroup{}
	done := make(chan struct{}, 10)
	for _, c := range cs {
		if ok := c.LoadCache(); ok {
			c.Logging(wc)
			continue
		}
		wg.Add(1)
		go func(ctx context.Context, c component.Component) {
			done <- struct{}{}
			c = h.SyncWithSource(c, ctx)
			c.StoreCache()
			c.Logging(wc)
			<-done
			wg.Done()
		}(ctx, c)
	}
	wg.Wait()
}

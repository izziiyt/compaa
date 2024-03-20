package handler

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/fatih/color"
	"github.com/izziiyt/compaa/component"
)

type Dockerfile struct {
	wc *component.WarnCondition
}

func NewDockerfile(wc *component.WarnCondition) *Dockerfile {
	return &Dockerfile{
		wc,
	}
}

func (h *Dockerfile) Handle(ctx context.Context, path string) {
	fmt.Printf("%v\n", path)

	cs, err := h.LookUp(path)
	if err != nil {
		color.Red("├ LookUp error: %v\n", err)
	}

	wg := &sync.WaitGroup{}
	done := make(chan struct{}, 10)
	for _, c := range cs {
		if ok := c.LoadCache(); ok {
			c.Logging(h.wc)
			continue
		}
		wg.Add(1)
		go func(ctx context.Context, c component.Component) {
			done <- struct{}{}
			if ok := c.LoadCache(); !ok {
				switch v := c.(type) {
				case *component.Image:
					v = v.SyncWithRegistry(ctx)
					v.StoreCache()
				}
			}
			c.Logging(h.wc)
			<-done
			wg.Done()
		}(ctx, c)
	}
	wg.Wait()
}

func (h *Dockerfile) LookUp(path string) ([]component.Component, error) {
	var buf []component.Component
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "FROM") {
			tokens := strings.Split(line, " ")
			if len(tokens) < 2 {
				continue
			}
			c := &component.Image{}
			c.FromRawString(tokens[1])
			buf = append(buf, c)
		}
	}
	return buf, nil
}

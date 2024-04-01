package handler

import (
	"bufio"
	"context"
	"os"
	"strings"

	"github.com/izziiyt/compaa/component"
)

type Dockerfile struct{}

func (h *Dockerfile) LookUp(path string) (buf []component.Component, err error) {
	f, err := os.Open(path)
	if err != nil {
		return
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
	return
}

func (h *Dockerfile) SyncWithSource(c component.Component, ctx context.Context) component.Component {
	switch v := c.(type) {
	case *component.Image:
		v = v.SyncWithRegistry(ctx)
		return v
	default:
		return v
	}
}

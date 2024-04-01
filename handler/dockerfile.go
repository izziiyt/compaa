package handler

import (
	"bufio"
	"context"
	"os"
	"strings"

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

func (h *Dockerfile) SyncWithSource(ctx context.Context, c component.Component) component.Component {
	switch v := c.(type) {
	case *component.Image:
		v = v.SyncWithRegistry(ctx)
		return v
	default:
		return v
	}
}

package handler

import (
	"bufio"
	"context"
	"os"
	"strings"

	"github.com/google/go-github/v60/github"
	"github.com/izziiyt/compaa/component"
)

type RequirementsTXT struct {
	gcli *github.Client
}

func NewRequirementsTXT(gcli *github.Client) *RequirementsTXT {
	return &RequirementsTXT{
		gcli,
	}
}

func (h *RequirementsTXT) LookUp(path string) ([]component.Component, error) {
	var buf []component.Component
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "-r") {
			continue
		}
		if strings.HasPrefix(line, "-c") {
			continue
		}
		if strings.HasPrefix(line, "./") {
			continue
		}
		if strings.HasPrefix(line, "https://") {
			continue
		}
		tokens := strings.Split(line, " ")
		tokens = strings.Split(tokens[0], "==")
		c := &component.Module{}
		c.Name = tokens[0]
		buf = append(buf, c)
	}

	return buf, nil
}

func (h *RequirementsTXT) SyncWithSource(ctx context.Context, c component.Component) component.Component {
	switch v := c.(type) {
	case *component.Module:
		v = v.SyncWithPypi(ctx)
		v = v.SyncWithGitHub(ctx, h.gcli)
		return v
	default:
		return v
	}
}

package handler

import (
	"bufio"
	"context"
	"os"
	"regexp"

	"github.com/google/go-github/v60/github"
	"github.com/izziiyt/compaa/component"
)

var (
	moduleRegexp   = regexp.MustCompile(`^\s*gem ['"]([^,'"]+)['"]`)
	languageRegexp = regexp.MustCompile(`^\s*ruby ['"](.+)['"]`)
)

type GemFile struct {
	gcli *github.Client
}

func NewGemFile(gcli *github.Client) *GemFile {
	return &GemFile{
		gcli,
	}
}

func (h *GemFile) LookUp(path string) ([]component.Component, error) {
	var buf []component.Component
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if match := languageRegexp.FindStringSubmatch(line); len(match) > 1 {
			c := &component.Language{}
			c.Name = "ruby"
			c.Version = string(match[1])
			buf = append(buf, c)
			continue
		}
		if match := moduleRegexp.FindStringSubmatch(line); len(match) > 1 {
			c := &component.Module{}
			c.Name = string(match[1])
			buf = append(buf, c)
			continue
		}
	}

	return buf, nil
}

func (h *GemFile) SyncWithSource(ctx context.Context, c component.Component) component.Component {
	switch v := c.(type) {
	case *component.Module:
		v = v.SyncWithRubyGem(ctx)
		v = v.SyncWithGitHub(ctx, h.gcli)
		return v
	case *component.Language:
		v = v.SyncWithEndOfLife(ctx)
		return v
	default:
		return v
	}
}

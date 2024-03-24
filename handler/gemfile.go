package handler

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"regexp"
	"sync"

	"github.com/fatih/color"
	"github.com/google/go-github/v60/github"
	"github.com/izziiyt/compaa/component"
)

var (
	moduleRegexp   = regexp.MustCompile(`^\s*gem ['"]([^,'"]+)['"]`)
	languageRegexp = regexp.MustCompile(`^\s*ruby ['"](.+)['"]`)
)

type GemFile struct {
	wc   *component.WarnCondition
	gcli *github.Client
}

func NewGemFile(wc *component.WarnCondition, gcli *github.Client) *GemFile {
	return &GemFile{
		wc,
		gcli,
	}
}

func (h *GemFile) Handle(ctx context.Context, path string) {
	fmt.Printf("%v\n", path)

	ts, err := h.LookUp(path)
	if err != nil {
		color.Red("LookUp error: %v", err)
		return
	}

	wg := &sync.WaitGroup{}
	done := make(chan struct{}, 10)
	for _, t := range ts {
		wg.Add(1)
		go func(ctx context.Context, t component.Component) {
			done <- struct{}{}
			if ok := t.LoadCache(); !ok {
				switch v := t.(type) {
				case *component.Module:
					v = v.SyncWithRubyGem(ctx)
					v = v.SyncWithGitHub(ctx, h.gcli)
					v.StoreCache()
				case *component.Language:
					v = v.SyncWithEndOfLife(ctx)
					v.StoreCache()
				}
			}
			t.StoreCache()
			t.Logging(h.wc)
			<-done
			wg.Done()
		}(ctx, t)
	}
	wg.Wait()
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

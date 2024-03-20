package handler

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/fatih/color"
	"github.com/google/go-github/v60/github"
	"github.com/izziiyt/compaa/component"
)

type RequirementsTXT struct {
	wc   *component.WarnCondition
	gcli *github.Client
}

func NewRequirementsTXT(wc *component.WarnCondition, gcli *github.Client) *RequirementsTXT {
	return &RequirementsTXT{
		wc,
		gcli,
	}
}

func (h *RequirementsTXT) Handle(ctx context.Context, path string) {
	fmt.Printf("%v\n", path)

	ts, err := h.LookUp(path)
	if err != nil {
		color.Red("├ LookUp error: %v\n", err)
	}

	wg := &sync.WaitGroup{}
	done := make(chan struct{}, 10)
	for _, t := range ts {
		if ok := t.LoadCache(); ok {
			t.Logging(h.wc)
			continue
		}
		wg.Add(1)
		go func(ctx context.Context, t component.Component) {
			done <- struct{}{}
			if ok := t.LoadCache(); !ok {
				switch v := t.(type) {
				case *component.Module:
					v = v.SyncWithPypi(ctx)
					v = v.SyncWithGitHub(ctx, h.gcli)
				}
			}
			t.Logging(h.wc)
			<-done
			wg.Done()
		}(ctx, t)
	}
	wg.Wait()
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

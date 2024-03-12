package main

import (
	"context"
	"depeol/component"
	"flag"
	"fmt"
	"os"
	"sync"
)

var (
	rd    = flag.Int("rd", 180, "recent days. used to determine log level")
	token = flag.String("t", "", "github token. recommend to set for sufficient github api rate limit")
)

func main() {
	flag.Parse()
	args := flag.Args()
	path := "."
	if len(args) > 0 {
		path = args[0]
	}
	entries, err := os.ReadDir(path)
	if err != nil {
		fmt.Println(err)
		return
	}
	ctx := context.Background()
	wc := &component.DefaultWarnCondition
	wc.RecentDays = *rd
	r := NewRouter(*token, wc)
	wg := &sync.WaitGroup{}
	done := make(chan struct{}, 10)
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		h := r.Route(e.Name())
		if h == nil {
			continue
		}
		wg.Add(1)
		go func(ctx context.Context, path string) {
			done <- struct{}{}
			h.Handle(ctx, path)
			<-done
			wg.Done()
		}(ctx, e.Name())
	}
	wg.Wait()
}

package component

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
	"github.com/izziiyt/compaa/sdk/dockerhub"
	"github.com/izziiyt/compaa/sdk/gcrio"
)

var (
	DockerHubRegistry          = "docker.io"
	GoogleContainerRegistry    = "gcr.io"
	DockerHubOfficialNamespace = "library"
	DefaultTag                 = "latest"
)

var imageCache = sync.Map{}

type Image struct {
	RawString  string
	Repository string
	Namespace  string
	Registry   string
	Tag        string
	Err        error
	LastUpdate time.Time
}

func (c *Image) FromRawString(s string) *Image {
	c.RawString = s
	tokens := strings.Split(s, "/")
	if len(tokens) > 3 {
		c.Err = fmt.Errorf("unsupported format: %v", s)
		return c
	}
	if len(tokens) == 1 { // golang:1.21.1-bullseye
		// official image
		c.Registry = DockerHubRegistry
		c.Namespace = DockerHubOfficialNamespace
		c.Repository = tokens[0]
	}
	if len(tokens) == 2 { // lstio/base
		c.Registry = DockerHubRegistry
		c.Namespace = tokens[0]
		c.Repository = tokens[1]
	}
	if len(tokens) == 3 { // gcr.io/distroless/base-nossl-debian11
		c.Registry = tokens[0]
		c.Namespace = tokens[1]
		c.Repository = tokens[2]
	}
	c.Tag = DefaultTag
	tokens = strings.Split(c.Repository, "@") // repository@sha256:xxxx
	c.Repository = tokens[0]
	tokens = strings.Split(c.Repository, ":") // repository:v1.2.3
	if len(tokens) > 1 {
		c.Repository = tokens[0]
		c.Tag = tokens[1]
	}
	return c
}

func (c *Image) Logging(wc *WarnCondition) {
	if c.Err != nil {
		color.Red("├ ERROR: %v %v\n", c.RawString, c.Err)
		return
	}
	if c.LastUpdate.AddDate(0, 0, wc.RecentDays).Before(time.Now()) {
		color.Yellow("├ WARN: %v last update isn't recent (%v)\n", c.RawString, c.LastUpdate)
		return
	}
	// color.Green("├ INFO: pass %v (%v)\n", c.RawString, c.LastUpdate)
}

func (c *Image) LoadCache() bool {
	v, ok := imageCache.Load(c.RawString)
	if ok {
		_v := v.(*Image)
		c.Repository = _v.Repository
		c.Namespace = _v.Namespace
		c.Registry = _v.Registry
		c.Tag = _v.Tag
		c.Err = _v.Err
		c.LastUpdate = _v.LastUpdate
	}
	return ok
}

func (c *Image) StoreCache() {
	imageCache.Store(c.RawString, c)
}

func (c *Image) SyncWithRegistry(ctx context.Context, cli *http.Client) *Image {
	if c.Err != nil {
		return c
	}
	if c.Registry == GoogleContainerRegistry {
		r, err := gcrio.ReadTag(ctx, cli, c.Namespace, c.Repository, c.Tag)
		if err != nil {
			c.Err = err
			return c
		}
		c.LastUpdate = r.Uploaded
		return c
	}
	if c.Registry == DockerHubRegistry {
		r, err := dockerhub.ReadTag(ctx, cli, c.Namespace, c.Repository, c.Tag)
		if err != nil {
			c.Err = err
			return c
		}
		c.LastUpdate = r.LastUpdated
		return c
	}

	c.Err = fmt.Errorf("unsupported registry: %s", c.Registry)
	return c
}

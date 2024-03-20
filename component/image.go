package component

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/izziiyt/compaa/sdk/dockerhub"
)

var (
	DockerHubRegistry          = "docker.io"
	DockerHubOfficialNamespace = "library"
	DefaultTag                 = "latest"
)

type Image struct {
	Repository string
	Namespace  string
	Registry   string
	Tag        string
	Err        error
	LastUpdate time.Time
}

func (c *Image) RawString() string {
	return fmt.Sprintf("%s/%s/%s:%s", c.Registry, c.Namespace, c.Repository, c.Tag)
}

func (c *Image) FromRawString(s string) *Image {
	tokens := strings.Split(s, "/")
	if len(tokens) > 3 {
		c.Err = fmt.Errorf("unexpected format: %v", s)
		return c
	}
	if len(tokens) == 1 {
		// official image
		c.Registry = DockerHubRegistry
		c.Namespace = DockerHubOfficialNamespace
		c.Repository = tokens[0]
	}
	if len(tokens) == 2 {
		c.Registry = DockerHubRegistry
		c.Namespace = tokens[0]
		c.Repository = tokens[1]
	}
	if len(tokens) == 3 {
		c.Registry = tokens[0]
		c.Namespace = tokens[1]
		c.Repository = tokens[2]
	}
	c.Tag = DefaultTag
	tokens = strings.Split(c.Repository, ":")
	if len(tokens) > 1 {
		c.Repository = tokens[0]
		c.Tag = tokens[1]
	}
	return c
}

var imageCache sync.Map

func init() {
	imageCache = sync.Map{}
}

func (c *Image) Logging(wc *WarnCondition) error {
	if c.Err != nil {
		_, err := fmt.Printf("├ ERROR: %v %v\n", c.RawString(), c.Err)
		return err
	}
	if c.LastUpdate.AddDate(0, 0, wc.RecentDays).Before(time.Now()) {
		_, err := fmt.Printf("├ WARN: %v last update isn't recent (%v)\n", c.RawString(), c.LastUpdate)
		return err
	}
	return nil
}

func (c *Image) LoadCache() bool {
	v, ok := imageCache.Load(c.RawString())
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
	imageCache.Store(c.RawString(), c)
}

func (c *Image) SyncWithDockerHub(ctx context.Context) *Image {
	if c.Registry != DockerHubRegistry {
		c.Err = fmt.Errorf("unsupported registry: %s", c.Registry)
		return c
	}
	r, err := dockerhub.ReadTag(ctx, c.Namespace, c.Repository, c.Tag)
	if err != nil {
		c.Err = err
		return c
	}
	c.LastUpdate = r.LastUpdated
	return c
}

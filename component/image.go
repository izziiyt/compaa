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

type RegistryHandler interface {
	ReadTag(ctx context.Context, cli *http.Client, namespace, repository, tag string) (time.Time, error)
}

var registryHandlers = map[string]RegistryHandler{
	GoogleContainerRegistry: &gcrioHandler{},
	DockerHubRegistry:       &dockerhubHandler{},
}

type gcrioHandler struct{}

func (h *gcrioHandler) ReadTag(ctx context.Context, cli *http.Client, namespace, repository, tag string) (time.Time, error) {
	r, err := gcrio.ReadTag(ctx, cli, namespace, repository, tag)
	if err != nil {
		return time.Time{}, err
	}
	return r.Uploaded, nil
}

type dockerhubHandler struct{}

func (h *dockerhubHandler) ReadTag(ctx context.Context, cli *http.Client, namespace, repository, tag string) (time.Time, error) {
	r, err := dockerhub.ReadTag(ctx, cli, namespace, repository, tag)
	if err != nil {
		return time.Time{}, err
	}
	return r.LastUpdated, nil
}

func (c *Image) FromRawString(s string) *Image {
	c.RawString = s
	parts := strings.Split(s, "/")
	switch len(parts) {
	case 1: // golang:1.21.1-bullseye or golang@sha256:xxxx
		c.Registry = DockerHubRegistry
		c.Namespace = DockerHubOfficialNamespace
		c.Repository = parts[0]
	case 2: // lstio/base or lstio/base:tag or lstio/base@sha256:xxxx
		c.Registry = DockerHubRegistry
		c.Namespace = parts[0]
		c.Repository = parts[1]
	case 3: // gcr.io/distroless/base-nossl-debian11 or gcr.io/distroless/base-nossl-debian11:tag or gcr.io/distroless/base-nossl-debian11@sha256:xxxx
		c.Registry = parts[0]
		c.Namespace = parts[1]
		c.Repository = parts[2]
	default:
		c.Err = fmt.Errorf("unsupported format: %v", s)
		return c
	}

	// Handle tag or digest
	repoParts := strings.Split(c.Repository, "@")
	if len(repoParts) == 2 {
		c.Repository = repoParts[0]
		c.Tag = "@" + repoParts[1] // Keep the digest format
	} else {
		repoParts = strings.Split(c.Repository, ":")
		if len(repoParts) == 2 {
			c.Repository = repoParts[0]
			c.Tag = repoParts[1]
		} else {
			c.Tag = DefaultTag
		}
	}

	return c
}

type Logger interface {
	Error(format string, a ...interface{})
	Warn(format string, a ...interface{})
	Info(format string, a ...interface{})
}

type DefaultLogger struct{}

func (l *DefaultLogger) Error(format string, a ...interface{}) {
	color.Red(format, a...)
}

func (l *DefaultLogger) Warn(format string, a ...interface{}) {
	color.Yellow(format, a...)
}

func (l *DefaultLogger) Info(format string, a ...interface{}) {
	color.Green(format, a...)
}

func (c *Image) Logging(wc *WarnCondition, logger Logger) {
	if logger == nil {
		logger = &DefaultLogger{}
	}

	if c.Err != nil {
		logger.Error("├ ERROR: %v %v\n", c.RawString, c.Err)
		return
	}
	if c.LastUpdate.AddDate(0, 0, wc.RecentDays).Before(time.Now()) {
		logger.Warn("├ WARN: %v last update isn't recent (%v)\n", c.RawString, c.LastUpdate.Format("2006-01-02"))
		return
	}
	// logger.Info("├ INFO: pass %v (%v)\n", c.RawString, c.LastUpdate)
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

	handler, ok := registryHandlers[c.Registry]
	if !ok {
		c.Err = fmt.Errorf("unsupported registry: %s", c.Registry)
		return c
	}

	lastUpdate, err := handler.ReadTag(ctx, cli, c.Namespace, c.Repository, c.Tag)
	if err != nil {
		c.Err = err
		return c
	}
	c.LastUpdate = lastUpdate

	return c
}

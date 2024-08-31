package main

import (
	"github.com/google/go-github/v60/github"
	"github.com/izziiyt/compaa/handler"
	"net/http"
	"strings"
)

type Router struct {
	gomod           *handler.GoMod
	packagejson     *handler.PackageJSON
	dockerfile      *handler.Dockerfile
	requirementstxt *handler.RequirementsTXT
	gemfile         *handler.GemFile
}

func NewRouter(ghtoken string, transport http.RoundTripper) *Router {
	hcli := &http.Client{
		Transport: transport,
	}
	gcli := github.NewClient(hcli)
	if ghtoken != "" {
		gcli = gcli.WithAuthToken(ghtoken)
	}
	return &Router{
		gomod:           &handler.GoMod{GCli: gcli},
		packagejson:     &handler.PackageJSON{GCli: gcli},
		dockerfile:      &handler.Dockerfile{},
		requirementstxt: &handler.RequirementsTXT{GCli: gcli},
		gemfile:         &handler.GemFile{GCli: gcli},
	}
}

func (r *Router) Route(path string) handler.Handler {
	path = strings.ToLower(path)
	if strings.Contains(path, "go.mod") {
		return r.gomod
	}
	if strings.Contains(path, "package.json") {
		return r.packagejson
	}
	if strings.Contains(path, "dockerfile") {
		return r.dockerfile
	}
	if strings.Contains(path, "requirements") && strings.Contains(path, ".txt") {
		return r.requirementstxt
	}
	if strings.Contains(path, "gemfile") {
		return r.gemfile
	}
	return nil
}

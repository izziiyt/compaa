package main

import (
	"strings"

	"github.com/google/go-github/v60/github"
	"github.com/izziiyt/compaa/component"
	"github.com/izziiyt/compaa/handler"
)

type Router struct {
	gomod           *handler.GoMod
	packagejson     *handler.PackageJSON
	dockerfile      *handler.Dockerfile
	requirementstxt *handler.RequirementsTXT
	gemfile         *handler.GemFile
}

func NewRouter(ghtoken string, wc *component.WarnCondition) *Router {
	gcli := github.NewClient(nil)
	if ghtoken != "" {
		gcli = gcli.WithAuthToken(ghtoken)
	}
	return &Router{
		gomod:           handler.NewGoMod(wc, gcli),
		packagejson:     handler.NewPackageJSON(wc, gcli),
		dockerfile:      handler.NewDockerfile(wc),
		requirementstxt: handler.NewRequirementsTXT(wc, gcli),
		gemfile:         handler.NewGemFile(wc, gcli),
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

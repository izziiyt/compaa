package main

import (
	"depeol/component"
	"depeol/eol"
	"depeol/handler"
	"depeol/npm"
	"github.com/google/go-github/v60/github"
	"strings"
)

type Router struct {
	gomod       *handler.GoMod
	packagejson *handler.PackageJSON
}

func NewRouter(ghtoken string, wc *component.WarnCondition) *Router {
	gcli := github.NewClient(nil)
	if ghtoken != "" {
		gcli = gcli.WithAuthToken(ghtoken)
	}
	ecli := eol.NewClient(nil)
	ncli := npm.NewClient(nil)
	return &Router{
		gomod:       handler.NewGoMod(gcli, ecli, wc),
		packagejson: handler.NewPackageJSON(gcli, ecli, ncli, wc),
	}
}

func (r *Router) Route(path string) handler.Handler {
	if strings.Contains(path, "go.mod") {
		return r.gomod
	}
	if strings.Contains(path, "package.json") {
		return r.packagejson
	}
	return nil
}

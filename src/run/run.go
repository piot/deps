package depsrun

import (
	"github.com/piot/deps/src/ccompile"
	"github.com/piot/deps/src/depsbuild"
	"github.com/piot/deps/src/depslib"
	"github.com/piot/log-go/src/clog"
)

func Run(info *depslib.DependencyInfo, log *clog.Log) error {
	artifacts, err := ccompile.Build(info, log)
	if err != nil {
		return err
	}
	primaryArtifact := artifacts[0]
	return depsbuild.Execute(log, primaryArtifact)
}

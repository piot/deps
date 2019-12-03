package command

import (
	"github.com/piot/deps/src/ccompile"
	"github.com/piot/deps/src/depslib"
	"github.com/piot/deps/src/depsrun"
	"github.com/piot/log-go/src/clog"
)

type Options struct {
	UseSymlink bool
	Artifact   depslib.ArtifactType
}

func setupDependencies(foundConfs []string, options Options, log *clog.Log) (*depslib.DependencyInfo, error) {
	dependencyInfo, err := depslib.SetupDependencies(foundConfs[0], options.UseSymlink, log)
	return dependencyInfo, err
}

func Build(foundConfs []string, options Options, log *clog.Log) error {
	dependencyInfo, depsErr := setupDependencies(foundConfs, options, log)
	if depsErr != nil {
		return depsErr
	}
	_, buildErr := ccompile.Build(dependencyInfo, options.Artifact, log)
	return buildErr
}

func Run(foundConfs []string, options Options, runArgs []string, log *clog.Log) error {
	dependencyInfo, depsErr := setupDependencies(foundConfs, options, log)
	if depsErr != nil {
		return depsErr
	}
	return depsrun.Run(dependencyInfo, options.Artifact, runArgs, log)
}

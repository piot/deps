/*---------------------------------------------------------------------------------------------
 *  Copyright (c) Peter Bjorklund. All rights reserved.
 *  Licensed under the MIT License. See LICENSE in the project root for license information.
 *--------------------------------------------------------------------------------------------*/

package command

import (
	"github.com/piot/deps/src/ccompile"
	"github.com/piot/deps/src/depslib"
	"github.com/piot/deps/src/depsrun"
)

type Options struct {
	Mode       depslib.Mode
	ForceClean bool
	Artifact   depslib.ArtifactType
}

func setupDependencies(foundConfs []string, options Options) (*depslib.DependencyInfo, error) {
	dependencyInfo, err := depslib.SetupDependencies(foundConfs[0], options.Mode, options.ForceClean)
	return dependencyInfo, err
}

func Build(foundConfs []string, options Options) error {
	dependencyInfo, depsErr := setupDependencies(foundConfs, options)
	if depsErr != nil {
		return depsErr
	}
	_, buildErr := ccompile.Build(dependencyInfo, options.Artifact)
	return buildErr
}

func Run(foundConfs []string, options Options, runArgs []string) error {
	dependencyInfo, depsErr := setupDependencies(foundConfs, options)
	if depsErr != nil {
		return depsErr
	}
	return depsrun.Run(dependencyInfo, options.Artifact, runArgs)
}

func Fetch(foundConfs []string, options Options) error {
	_, depsErr := setupDependencies(foundConfs, options)
	if depsErr != nil {
		return depsErr
	}
	return nil
}

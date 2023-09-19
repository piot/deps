/*---------------------------------------------------------------------------------------------
 *  Copyright (c) Peter Bjorklund. All rights reserved.
 *  Licensed under the MIT License. See LICENSE in the project root for license information.
 *--------------------------------------------------------------------------------------------*/

package command

import (
	"github.com/piot/deps/src/depslib"
)

type Options struct {
	Mode                       depslib.Mode
	ForceClean                 bool
	UseDevelopmentDependencies bool
	Artifact                   depslib.ArtifactType
	LocalPackageRoot           string
	TargetDepsPath             string
}

func setupDependencies(foundConfs []string, options Options) (*depslib.DependencyInfo, error) {
	dependencyInfo, err := depslib.SetupDependencies(foundConfs[0], options.Mode, options.ForceClean, options.UseDevelopmentDependencies, options.LocalPackageRoot, options.TargetDepsPath)
	return dependencyInfo, err
}

func Fetch(foundConfs []string, options Options, showTree bool) error {
	root, depsErr := setupDependencies(foundConfs, options)
	if depsErr != nil {
		return depsErr
	}

	if showTree {
		root.RootNode.Print(0)
	}

	return nil
}

/*---------------------------------------------------------------------------------------------
 *  Copyright (c) Peter Bjorklund. All rights reserved.
 *  Licensed under the MIT License. See LICENSE in the project root for license information.
 *--------------------------------------------------------------------------------------------*/

package depsrun

import (
	"github.com/piot/deps/src/ccompile"
	"github.com/piot/deps/src/depsbuild"
	"github.com/piot/deps/src/depslib"
)

func Run(info *depslib.DependencyInfo, override depslib.ArtifactType, runArgs []string) error {
	artifacts, err := ccompile.Build(info, override)
	if err != nil {
		return err
	}

	primaryArtifact := artifacts[0]

	return depsbuild.Execute(primaryArtifact, runArgs...)
}

/*---------------------------------------------------------------------------------------------
 *  Copyright (c) Peter Bjorklund. All rights reserved.
 *  Licensed under the MIT License. See LICENSE in the project root for license information.
 *--------------------------------------------------------------------------------------------*/

package main

import (
	"fmt"
	"os"

	"github.com/alecthomas/kong"
	"github.com/piot/deps/src/command"
	"github.com/piot/deps/src/depslib"
)

var version string

// SharedOptions are command line shared options.
type SharedOptions struct {
	Mode                       string `name:"mode" short:"m" enum:"wget,symlink,clone,read" default:"wget" help:"How the dependencies are realized: wget, symlink or clone"`
	ForceClean                 bool   `name:"clean" default:"false" help:"delete the deps directory"`
	LocalPackageRoot           string `name:"localPackageRoot" short:"r" default:"" type:"path" help:"root directory of local packages"`
	TargetDepsPath             string `name:"targetDepsPath" short:"t" default:"" type:"path" help:"deps/ target directory"`
	UseDevelopmentDependencies bool   `name:"dev" default:"false" help:"include the development dependencies"`
	Artifact                   string `short:"a" optional:"" help:"override application type"`
}

// FetchCmd is the options for a fetch.
type FetchCmd struct {
	Shared   SharedOptions `embed:""`
	ShowTree bool          `name:"tree" default:"false" help:"show the dependency tree"`
}

// Options are all the command line options.
type Options struct {
	Fetch FetchCmd `cmd:""`
}

func stringToArtifactType(appType string) depslib.ArtifactType {
	switch appType {
	case "application":
		return depslib.Application
	case "console":
		return depslib.ConsoleApplication
	case "library":
		return depslib.Library
	}

	return depslib.Inherit
}

func sharedOptionsToGeneralOptions(shared SharedOptions) command.Options {
	mode := depslib.Wget

	switch shared.Mode {
	case "wget":
		mode = depslib.Wget
	case "symlink":
		mode = depslib.Symlink
	case "clone":
		mode = depslib.Clone
	case "read":
		mode = depslib.ReadLocal
	}

	generalOptions := command.Options{Mode: mode, ForceClean: shared.ForceClean,
		UseDevelopmentDependencies: shared.UseDevelopmentDependencies, LocalPackageRoot: shared.LocalPackageRoot,
		TargetDepsPath: shared.TargetDepsPath, Artifact: stringToArtifactType(shared.Artifact)}

	return generalOptions
}

// Run is called if a fetch command was issued.
func (o *FetchCmd) Run() error {
	foundConfs, foundErr := depslib.FindClosestConfigurationFiles(".")
	if foundErr != nil {
		return foundErr
	}

	return command.Fetch(foundConfs, sharedOptionsToGeneralOptions(o.Shared), o.ShowTree)
}

func main() {
	ctx := kong.Parse(&Options{})

	err := ctx.Run()
	if err != nil {
		fmt.Printf("ERROR:%v\n", err)
		os.Exit(-1)
	}
}

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
	Mode       string `name:"mode" short:"m" enum:"wget,symlink,clone" default:"wget" help:"How the dependencies are realized: wget, symlink or clone"`
	ForceClean bool   `name:"clean" default:"false" help:"delete the deps directory"`
	Artifact   string `short:"a" optional:"" help:"override application type"`
}

// BuildCmd is the options for a build.
type BuildCmd struct {
	Shared SharedOptions `embed:""`
}

// RunCmd is the options for a run command.
type RunCmd struct {
	Shared  SharedOptions `embed:""`
	RunArgs []string      `arg:"" help:"run arguments"`
}

// FetchCmd is the options for a fetch.
type FetchCmd struct {
	Shared SharedOptions `embed:""`
}

// Options are all the command line options.
type Options struct {
	Fetch FetchCmd `cmd:""`
	Build BuildCmd `cmd:""`
	Run   RunCmd   `cmd:""`
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
	}

	generalOptions := command.Options{Mode: mode, ForceClean: shared.ForceClean, Artifact: stringToArtifactType(shared.Artifact)}

	return generalOptions
}

// Run is called if a run command was issued.
func (o *RunCmd) Run() error {
	foundConfs, foundErr := depslib.FindClosestConfigurationFiles(".")
	if foundErr != nil {
		return foundErr
	}

	return command.Run(foundConfs, sharedOptionsToGeneralOptions(o.Shared), o.RunArgs)
}

// Run is called if a build command was issued.
func (o *BuildCmd) Run() error {
	foundConfs, foundErr := depslib.FindClosestConfigurationFiles(".")
	if foundErr != nil {
		return foundErr
	}

	return command.Build(foundConfs, sharedOptionsToGeneralOptions(o.Shared))
}

// Run is called if a fetch command was issued.
func (o *FetchCmd) Run() error {
	foundConfs, foundErr := depslib.FindClosestConfigurationFiles(".")
	if foundErr != nil {
		return foundErr
	}

	return command.Fetch(foundConfs, sharedOptionsToGeneralOptions(o.Shared))
}

func main() {
	ctx := kong.Parse(&Options{})

	err := ctx.Run()
	if err != nil {
		fmt.Printf("ERROR:%v\n", err)
		os.Exit(-1)
	}
}

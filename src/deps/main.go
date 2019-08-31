/*

MIT License

Copyright (c) 2019 Peter Bjorklund

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.

*/

package main

import (
	"github.com/piot/cli-go/src/cli"
	"github.com/piot/deps/src/command"
	"github.com/piot/deps/src/depslib"
	"github.com/piot/log-go/src/clog"
)

var Version string

type SharedOptions struct {
	Symlink  bool   `name:"symlink" short:"l"  help:"symlink using the parent directory instead of downloading"`
	Artifact string `short:"a" optional:"" help:"override application type"`
}

type BuildCmd struct {
	Shared SharedOptions `embed:""`
}

type RunCmd struct {
	Shared SharedOptions `embed:""`
}

type Options struct {
	Build   BuildCmd    `cmd:""`
	Run     RunCmd      `cmd:""`
	Options cli.Options `embed:""`
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
	generalOptions := command.Options{UseSymlink: shared.Symlink, Artifact: stringToArtifactType(shared.Artifact)}
	return generalOptions
}

func (o *RunCmd) Run(log *clog.Log) error {
	foundConfs, foundErr := depslib.FindClosestConfigurationFiles(".", log)
	if foundErr != nil {
		return foundErr
	}
	return command.Run(foundConfs, sharedOptionsToGeneralOptions(o.Shared), log)
}

func (o *BuildCmd) Run(log *clog.Log) error {
	foundConfs, foundErr := depslib.FindClosestConfigurationFiles(".", log)
	if foundErr != nil {
		return foundErr
	}
	return command.Build(foundConfs, sharedOptionsToGeneralOptions(o.Shared), log)
}

func main() {
	cli.Run(&Options{}, cli.RunOptions{Version: Version, ApplicationType: cli.Utility}, nil)
}

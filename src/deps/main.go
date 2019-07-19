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
	"flag"
	"os"

	"github.com/piot/deps/src/ccompile"
	"github.com/piot/deps/src/depslib"
	depsrun "github.com/piot/deps/src/run"
	"github.com/piot/log-go/src/clog"
	"github.com/piot/log-go/src/clogint"
)

func run(log *clog.Log) error {
	foundConfs, foundErr := depslib.FindClosestConfigurationFiles(".", log)
	if foundErr != nil {
		return foundErr
	}
	dependencyInfo, err := depslib.SetupDependencies(foundConfs[0], log)
	if err != nil {
		return err
	}
	if len(os.Args) >= 2 {
		cmd := os.Args[1]
		if cmd == "build" {
			override := depslib.Inherit
			if len(os.Args) >= 3 {
				artifact := os.Args[2]
				if artifact == "console" {
					override = depslib.ConsoleApplication
				}
			}
			_, buildErr := ccompile.Build(dependencyInfo, override, log)
			return buildErr
		} else if cmd == "run" {
			return depsrun.Run(dependencyInfo, depslib.Inherit, log)
		}
	}
	useSymlink := flag.Bool("l", false, "use local symlink instead of download")
	flag.Parse()
	return depslib.Install(dependencyInfo, *useSymlink, log)
}

func main() {
	log := clog.DefaultLog()
	log.SetLogLevel(clogint.Debug)
	log.Info("deps")
	err := run(log)
	if err != nil {
		log.Err(err)
		os.Exit(1)
	}
}

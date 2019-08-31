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
	"fmt"
	"os"

	"github.com/piot/deps/src/command"
	"github.com/piot/deps/src/depslib"
	"github.com/piot/log-go/src/clog"
	"github.com/piot/log-go/src/clogint"
)

func run(log *clog.Log) error {
	foundConfs, foundErr := depslib.FindClosestConfigurationFiles(".", log)
	if foundErr != nil {
		return foundErr
	}

	if len(os.Args) < 2 {
		return fmt.Errorf("subcommand is required")
	}
	general := flag.NewFlagSet("default", flag.ExitOnError)
	useSymlink := general.Bool("l", false, "use local symlink instead of download")
	artifactType := general.String("a", "application", "artifact")
	general.Parse(os.Args[2:])
	o := command.Options{}
	o.UseSymlink = *useSymlink
	o.Artifact = depslib.Inherit
	if *artifactType == "console" {
		o.Artifact = depslib.ConsoleApplication
	}

	cmd := os.Args[1]
	switch cmd {
	case "build":
		return command.Build(foundConfs, o, log)
	case "run":
		return command.Run(foundConfs, o, log)
	}
	return nil
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
	log.Info("done")
}

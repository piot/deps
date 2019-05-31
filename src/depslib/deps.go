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

package depslib

import (
	"path"
	"strings"

	"github.com/piot/log-go/src/clog"
)

func HackRemoveCShortName(shortname string) string {
	if strings.HasSuffix(shortname, "-c") {
		return shortname[:len(shortname)-2]
	}
	return shortname
}

func copyDependency(rootPath string, depsPath string, repoName string, log *clog.Log) error {
	shortName := RepoNameToShortName(repoName)
	includeShortName := HackRemoveCShortName(shortName)
	packageDir := path.Join(rootPath, "../", shortName+"/")
	targetName := path.Join(depsPath, shortName)
	log.Debug("installing", clog.String("packageName", repoName), clog.String("shortName", shortName), clog.String("target", targetName))
	makeErr := MakeRelativeSymlink(packageDir, targetName, log)
	if makeErr != nil {
		return makeErr
	}
	sourceInclude := path.Join(packageDir, "src", "include", includeShortName)
	targetInclude := path.Join(depsPath, "include", includeShortName)
	includeErr := MakeRelativeSymlink(sourceInclude, targetInclude, log)
	return includeErr
}

func SetupDependencies(filename string, log *clog.Log) error {
	conf, confErr := ReadConfigFromFilename(filename)
	if confErr != nil {
		return confErr
	}
	rootPath := path.Dir(filename)
	depsPath := path.Join(rootPath, "deps/")
	for _, dep := range conf.Dependencies {
		copyErr := copyDependency(rootPath, depsPath, dep.Name, log)
		if copyErr != nil {
			return copyErr
		}
	}
	const useDevelopmentDependencies = true
	if useDevelopmentDependencies {
		for _, dep := range conf.Development {
			copyErr := copyDependency(rootPath, depsPath, dep.Name, log)
			if copyErr != nil {
				return copyErr
			}
		}
	}
	return nil
}

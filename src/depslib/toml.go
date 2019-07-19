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
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	toml "github.com/pelletier/go-toml"
	"github.com/piot/log-go/src/clog"
)

type Package struct {
	Version string
	Name    string
}

func (p Package) String() string {
	return fmt.Sprintf("name:%v version:%v", p.Name, p.Version)
}

type Config struct {
	DepsVersion  string
	Version      string
	Name         string
	ArtifactType string
	Dependencies []Package
	Development  []Package
}

func RepoNameToShortName(repo string) string {
	projectNames := strings.Split(repo, "/")
	if len(projectNames) != 2 {
		panic("wrong project name:" + repo)
	}
	return projectNames[1]
}

func ReadFromReader(reader io.Reader) (*Config, error) {
	tomlString, tomlParseErr := ioutil.ReadAll(reader)
	if tomlParseErr != nil {
		return nil, tomlParseErr
	}
	config := &Config{}
	unmarshalErr := toml.Unmarshal(tomlString, config)
	if unmarshalErr != nil {
		return nil, unmarshalErr
	}
	if config.DepsVersion != "0.0.0" {
		return nil, fmt.Errorf("wrong deps file format version '%v'", config.DepsVersion)
	}

	return config, unmarshalErr
}

func ReadConfigFromFilename(filename string) (*Config, error) {
	reader, openErr := os.Open(filename)
	if openErr != nil {
		return nil, openErr
	}
	return ReadFromReader(reader)
}

func ReadConfigFromDirectory(directory string, log *clog.Log) (*Config, error) {
	log.Debug("read config", clog.String("directory", directory))
	info, statErr := os.Stat(directory)
	if statErr != nil {
		return nil, statErr
	}

	if !info.IsDir() {
		return nil, fmt.Errorf("deps: not a directory %v", directory)
	}

	filename := filepath.Join(directory, "deps.toml")
	return ReadConfigFromFilename(filename)
}

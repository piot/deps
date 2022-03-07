/*---------------------------------------------------------------------------------------------
 *  Copyright (c) Peter Bjorklund. All rights reserved.
 *  Licensed under the MIT License. See LICENSE in the project root for license information.
 *--------------------------------------------------------------------------------------------*/

package depslib

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	toml "github.com/pelletier/go-toml"
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

	return repo
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

func ReadConfigFromDirectory(directory string) (*Config, error) {
	fmt.Printf("reading config from '%s'\n", directory)
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

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
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/blang/semver"

	"github.com/piot/log-go/src/clog"
)

func HackRemoveCShortName(shortname string) string {
	if strings.HasSuffix(shortname, "-c") {
		return shortname[:len(shortname)-2]
	}
	return shortname
}

func symlinkSrcInclude(packageDir string, depsPath string, shortName string, log *clog.Log) error {
	includeShortName := HackRemoveCShortName(shortName)
	sourceInclude := path.Join(packageDir, "src", "include", includeShortName)
	targetInclude := path.Join(depsPath, "include", includeShortName)
	includeErr := MakeRelativeSymlink(sourceInclude, targetInclude, log)
	return includeErr
}

func symlinkRepo(rootPath string, depsPath string, repoName string, log *clog.Log) error {
	shortName := RepoNameToShortName(repoName)
	packageDir := path.Join(rootPath, shortName+"/")
	targetName := path.Join(depsPath, shortName)
	log.Debug("installing", clog.String("packageName", repoName), clog.String("shortName", shortName), clog.String("target", targetName))
	makeErr := MakeRelativeSymlink(packageDir, targetName, log)
	if makeErr != nil {
		return makeErr
	}
	return symlinkSrcInclude(packageDir, depsPath, shortName, log)
}

func wgetRepo(rootPath string, depsPath string, repoName string, log *clog.Log) error {
	downloadURLString := fmt.Sprintf("https://github.com/%v/archive/master.zip", repoName)
	downloadURL, parseErr := url.Parse(downloadURLString)
	if parseErr != nil {
		return parseErr
	}
	contentReader, downloadErr := HTTPGet(downloadURL, log)
	if downloadErr != nil {
		return downloadErr
	}
	targetFile, createErr := os.Create("temp.zip")
	if createErr != nil {
		return createErr
	}
	_, copyErr := io.Copy(targetFile, contentReader)
	if copyErr != nil {
		return copyErr
	}
	contentReader.(io.Closer).Close()
	targetFile.Close()
	shortName := RepoNameToShortName(repoName)
	targetDirectory := path.Join(depsPath, shortName)
	zipPrefix := fmt.Sprintf("%v-master/", shortName)
	unzipErr := unzipFile("temp.zip", targetDirectory, zipPrefix, log)
	if unzipErr != nil {
		return unzipErr
	}
	return symlinkSrcInclude(targetDirectory, depsPath, shortName, log)
}

func copyDependency(rootPath string, depsPath string, repoName string, useSymlink bool, log *clog.Log) error {
	if useSymlink {
		return symlinkRepo(rootPath, depsPath, repoName, log)
	}

	return wgetRepo(rootPath, depsPath, repoName, log)
}

func establishPackageAndReadConfig(rootPath string, depsPath string, packageName string, useSymlink bool, log *clog.Log) (*Config, error) {
	log.Debug("establish package", clog.String("packageName", packageName), clog.String("depsPath", depsPath))
	copyErr := copyDependency(rootPath, depsPath, packageName, useSymlink, log)
	if copyErr != nil {
		return nil, copyErr
	}

	directoryName := RepoNameToShortName(packageName)
	packageDirectory := path.Join(depsPath, directoryName)
	conf, confErr := ReadConfigFromDirectory(packageDirectory, log)
	if confErr != nil {
		return nil, confErr
	}
	if conf.Name != packageName {
		return nil, fmt.Errorf("name mismatch %v vs %v", conf.Name, packageName)
	}
	return conf, confErr
}

type DependencyNode struct {
	name            string
	version         semver.Version
	artifactType    ArtifactType
	dependencies    []*DependencyNode
	development     []*DependencyNode
	dependingOnThis []*DependencyNode
}

func (n *DependencyNode) Name() string {
	return n.name
}

func (n *DependencyNode) ArtifactType() ArtifactType {
	return n.artifactType
}

func (n *DependencyNode) ShortName() string {
	return RepoNameToShortName(n.name)
}

func (n *DependencyNode) Dependencies() []*DependencyNode {
	return n.dependencies
}

func (n *DependencyNode) AddDependingOnThis(node *DependencyNode) {
	n.dependingOnThis = append(n.dependingOnThis, node)
}

func (n *DependencyNode) AddDependency(node *DependencyNode) {
	n.dependencies = append(n.dependencies, node)
	node.AddDependingOnThis(n)
}

func (n *DependencyNode) AddDevelopment(node *DependencyNode) {
	n.development = append(n.development, node)
}

func (n *DependencyNode) String() string {
	return fmt.Sprintf("node %v %v", n.name, n.version)
}

func (n *DependencyNode) Print(indent int) {
	indentString := strings.Repeat("..", indent)
	fmt.Printf("%s %v\n", indentString, n)

	for _, depNode := range n.dependencies {
		depNode.Print(indent + 1)
	}
}

type DependencyInfo struct {
	RootPath        string
	PackageRootPath string
	RootNodes       []*DependencyNode
	RootNode        *DependencyNode
}

type Cache struct {
	nodes map[string]*DependencyNode
}

func NewCache() *Cache {
	return &Cache{nodes: make(map[string]*DependencyNode)}
}

func (c *Cache) FindNode(name string) *DependencyNode {
	return c.nodes[name]
}
func (c *Cache) AddNode(name string, node *DependencyNode) {
	c.nodes[name] = node
}

func handleNode(rootPath string, depsPath string, node *DependencyNode, cache *Cache, depName string, useSymlink bool, log *clog.Log) (*DependencyNode, error) {
	foundNode := cache.FindNode(depName)
	if foundNode == nil {
		log.Trace("didnt find it, need to read", clog.String("depName", depName))
		depConf, confErr := establishPackageAndReadConfig(rootPath, depsPath, depName, useSymlink, log)
		if confErr != nil {
			return nil, confErr
		}
		var convertErr error
		foundNode, convertErr = convertFromConfigNode(rootPath, depsPath, depConf, cache, useSymlink, log)
		if convertErr != nil {
			return nil, convertErr
		}
	}
	return foundNode, nil
}

type ArtifactType uint

const (
	Library ArtifactType = iota
	ConsoleApplication
	Application
	Inherit
)

func ToArtifactType(v string) ArtifactType {
	if v == "lib" {
		return Library
	}
	if v == "console" {
		return ConsoleApplication
	}
	if v == "executable" {
		return Application
	}
	return Library
}

func convertFromConfigNode(rootPath string, depsPath string, conf *Config, cache *Cache, useSymlink bool, log *clog.Log) (*DependencyNode, error) {
	artifactType := ToArtifactType(conf.ArtifactType)
	node := &DependencyNode{name: conf.Name, version: semver.MustParse(conf.Version), artifactType: artifactType}
	cache.AddNode(conf.Name, node)
	for _, dep := range conf.Dependencies {
		foundNode, handleErr := handleNode(rootPath, depsPath, node, cache, dep.Name, useSymlink, log)
		if handleErr != nil {
			return nil, handleErr
		}
		node.AddDependency(foundNode)
	}
	const useDevelopmentDependencies = true
	if useDevelopmentDependencies {
		for _, dep := range conf.Development {
			_, handleErr := handleNode(rootPath, depsPath, node, cache, dep.Name, useSymlink, log)
			if handleErr != nil {
				return nil, handleErr
			}
			//node.AddDevelopment(foundNode)
		}
	}

	return node, nil
}

func calculateTotalDependencies(rootPath string, depsPath string, conf *Config, useSymlink bool, log *clog.Log) (*Cache, *DependencyNode, error) {

	cache := NewCache()
	rootNode, rootNodeErr := convertFromConfigNode(rootPath, depsPath, conf, cache, useSymlink, log)
	return cache, rootNode, rootNodeErr
}

func SetupDependencies(filename string, useSymlink bool, log *clog.Log) (*DependencyInfo, error) {
	log.Debug("setup dependencies", clog.String("configFilename", filename))
	conf, confErr := ReadConfigFromFilename(filename)
	if confErr != nil {
		return nil, confErr
	}
	packageRootPath := path.Dir(filename)
	rootPath := path.Dir(packageRootPath)
	depsPath := filepath.Join(packageRootPath, "deps/")
	BackupDeps(depsPath, log)
	log.Debug("calculate", clog.String("rootPath", rootPath), clog.String("packageRootPath", packageRootPath), clog.String("depsPath", depsPath))

	cache, rootNode, rootNodeErr := calculateTotalDependencies(rootPath, depsPath, conf, useSymlink, log)
	if rootNodeErr != nil {
		return nil, rootNodeErr
	}
	var rootNodes []*DependencyNode
	for _, node := range cache.nodes {
		if node.name == rootNode.name {
			continue
		}
		rootNodes = append(rootNodes, node)
	}

	//rootNode.Print(0)
	info := &DependencyInfo{RootPath: rootPath, PackageRootPath: packageRootPath, RootNode: rootNode, RootNodes: rootNodes}
	return info, nil
}

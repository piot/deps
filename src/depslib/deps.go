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

func copyDependency(rootPath string, depsPath string, repoName string, log *clog.Log) error {
	if false {
		return symlinkRepo(rootPath, depsPath, repoName, log)
	}

	return wgetRepo(rootPath, depsPath, repoName, log)
}

func readConfigFromLocalPackageName(rootPath string, packageName string, log *clog.Log) (*Config, error) {
	directoryName := RepoNameToShortName(packageName)
	packageDirectory := path.Join(rootPath, directoryName)
	conf, confErr := ReadConfigFromDirectory(packageDirectory)
	if conf.Name != packageName {
		return nil, fmt.Errorf("name mismatch %v vs %v", conf.Name, packageName)
	}
	return conf, confErr
}

type DependencyNode struct {
	name            string
	version         semver.Version
	dependencies    []*DependencyNode
	development     []*DependencyNode
	dependingOnThis []*DependencyNode
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

func install(rootPath string, packageRootPath string, nodes map[string]*DependencyNode, log *clog.Log) error {
	depsPath := path.Join(packageRootPath, "deps/")
	CleanDirectoryWithBackup(depsPath, "deps.clean", log)
	for _, dep := range nodes {
		copyErr := copyDependency(rootPath, depsPath, dep.name, log)
		if copyErr != nil {
			return copyErr
		}
	}
	return nil
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

func handleNode(rootPath string, node *DependencyNode, cache *Cache, depName string, log *clog.Log) (*DependencyNode, error) {
	foundNode := cache.FindNode(depName)
	if foundNode == nil {
		log.Info("didnt find it, need to read", clog.String("depName", depName))
		depConf, confErr := readConfigFromLocalPackageName(rootPath, depName, log)
		if confErr != nil {
			return nil, confErr
		}
		var convertErr error
		foundNode, convertErr = convertFromConfigNode(rootPath, depConf, cache, log)
		if convertErr != nil {
			return nil, convertErr
		}
	}
	return foundNode, nil
}

func convertFromConfigNode(rootPath string, conf *Config, cache *Cache, log *clog.Log) (*DependencyNode, error) {
	node := &DependencyNode{name: conf.Name, version: semver.MustParse(conf.Version)}
	cache.AddNode(conf.Name, node)
	for _, dep := range conf.Dependencies {
		foundNode, handleErr := handleNode(rootPath, node, cache, dep.Name, log)
		if handleErr != nil {
			return nil, handleErr
		}
		node.AddDependency(foundNode)
	}
	const useDevelopmentDependencies = true
	if useDevelopmentDependencies {
		for _, dep := range conf.Development {
			_, handleErr := handleNode(rootPath, node, cache, dep.Name, log)
			if handleErr != nil {
				return nil, handleErr
			}
			//node.AddDevelopment(foundNode)
		}
	}

	return node, nil
}

func calculateTotalDependencies(rootPath string, conf *Config, log *clog.Log) (*Cache, *DependencyNode, error) {
	cache := NewCache()
	rootNode, rootNodeErr := convertFromConfigNode(rootPath, conf, cache, log)
	return cache, rootNode, rootNodeErr
}

func SetupDependencies(filename string, log *clog.Log) error {
	conf, confErr := ReadConfigFromFilename(filename)
	if confErr != nil {
		return confErr
	}
	packageRootPath := path.Dir(filename)
	rootPath := path.Dir(packageRootPath)
	fmt.Printf("roootpath:%v\n", rootPath)
	cache, rootNode, rootNodeErr := calculateTotalDependencies(rootPath, conf, log)
	if rootNodeErr != nil {
		return rootNodeErr
	}
	rootNode.Print(0)
	return install(rootPath, packageRootPath, cache.nodes, log)
}

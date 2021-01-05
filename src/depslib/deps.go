/*---------------------------------------------------------------------------------------------
 *  Copyright (c) Peter Bjorklund. All rights reserved.
 *  Licensed under the MIT License. See LICENSE in the project root for license information.
 *--------------------------------------------------------------------------------------------*/

package depslib

import (
	"fmt"
	"io"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"github.com/blang/semver"
)

type Mode uint8

const (
	Wget Mode = iota
	Clone
	Symlink
)

func HackRemoveCShortName(shortname string) string {
	if strings.HasSuffix(shortname, "-c") {
		return shortname[:len(shortname)-2]
	}
	return shortname
}

func symlinkSrcInclude(packageDir string, depsPath string, shortName string) error {
	includeShortName := HackRemoveCShortName(shortName)
	sourceInclude := path.Join(packageDir, "src", "include", includeShortName)
	targetInclude := path.Join(depsPath, "include", includeShortName)
	includeErr := MakeRelativeSymlink(sourceInclude, targetInclude)
	return includeErr
}

func symlinkRepo(rootPath string, depsPath string, repoName string) error {
	shortName := RepoNameToShortName(repoName)
	packageDir := path.Join(rootPath, shortName+"/")
	targetName := path.Join(depsPath, shortName)
	makeErr := MakeRelativeSymlink(packageDir, targetName)
	if makeErr != nil {
		return makeErr
	}
	return symlinkSrcInclude(packageDir, depsPath, shortName)
}

func wgetRepo(rootPath string, depsPath string, repoName string) error {
	downloadURLString := fmt.Sprintf("https://github.com/%v/archive/master.zip", repoName)
	downloadURL, parseErr := url.Parse(downloadURLString)
	if parseErr != nil {
		return parseErr
	}
	contentReader, downloadErr := HTTPGet(downloadURL)
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
	unzipErr := unzipFile("temp.zip", targetDirectory, zipPrefix)
	if unzipErr != nil {
		return unzipErr
	}
	return symlinkSrcInclude(targetDirectory, depsPath, shortName)
}

func gitClone(depsPath string, repoName string) error {
	downloadURLString := fmt.Sprintf("https://github.com/%v.git", repoName)
	downloadURL, parseErr := url.Parse(downloadURLString)
	if parseErr != nil {
		return parseErr
	}

	cmd := exec.Command("git", "clone", downloadURL.String())

	cmd.Dir = depsPath

	cmd.Start()

	cmd.Wait()

	return nil
}

func gitPull(targetDirectory string, repoName string) error {
	cmd := exec.Command("git", "pull")

	cmd.Dir = targetDirectory

	cmd.Start()

	cmd.Wait()

	return nil
}

func directoryExists(directory string) bool {
	stat, checkDirectoryErr := os.Lstat(directory)

	return checkDirectoryErr == nil && stat.IsDir()
}

func cloneOrPullRepo(targetDirectory string, depsPath string, repoName string) error {
	if directoryExists(targetDirectory) {
		return gitPull(targetDirectory, repoName)
	}
	return gitClone(depsPath, repoName)

}

func copyDependency(rootPath string, depsPath string, repoName string, mode Mode) error {
	shortName := RepoNameToShortName(repoName)
	targetDirectory := path.Join(depsPath, shortName)
	switch mode {
	case Symlink:
		return symlinkRepo(rootPath, depsPath, repoName)
	case Clone:
		return cloneOrPullRepo(targetDirectory, depsPath, repoName)
	case Wget:
		return wgetRepo(rootPath, depsPath, repoName)
	default:
		return fmt.Errorf("unknown mode")
	}
}

func establishPackageAndReadConfig(rootPath string, depsPath string, packageName string, mode Mode) (*Config, error) {
	copyErr := copyDependency(rootPath, depsPath, packageName, mode)
	if copyErr != nil {
		return nil, copyErr
	}

	directoryName := RepoNameToShortName(packageName)
	packageDirectory := path.Join(depsPath, directoryName)
	conf, confErr := ReadConfigFromDirectory(packageDirectory)
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

func handleNode(rootPath string, depsPath string, node *DependencyNode, cache *Cache, depName string, mode Mode) (*DependencyNode, error) {
	foundNode := cache.FindNode(depName)
	if foundNode == nil {
		depConf, confErr := establishPackageAndReadConfig(rootPath, depsPath, depName, mode)
		if confErr != nil {
			return nil, confErr
		}
		var convertErr error
		foundNode, convertErr = convertFromConfigNode(rootPath, depsPath, depConf, cache, mode)
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

func convertFromConfigNode(rootPath string, depsPath string, conf *Config, cache *Cache, mode Mode) (*DependencyNode, error) {
	artifactType := ToArtifactType(conf.ArtifactType)
	node := &DependencyNode{name: conf.Name, version: semver.MustParse(conf.Version), artifactType: artifactType}
	cache.AddNode(conf.Name, node)
	for _, dep := range conf.Dependencies {
		foundNode, handleErr := handleNode(rootPath, depsPath, node, cache, dep.Name, mode)
		if handleErr != nil {
			return nil, handleErr
		}
		node.AddDependency(foundNode)
	}
	const useDevelopmentDependencies = true
	if useDevelopmentDependencies {
		for _, dep := range conf.Development {
			_, handleErr := handleNode(rootPath, depsPath, node, cache, dep.Name, mode)
			if handleErr != nil {
				return nil, handleErr
			}
			//node.AddDevelopment(foundNode)
		}
	}

	return node, nil
}

func calculateTotalDependencies(rootPath string, depsPath string, conf *Config, mode Mode) (*Cache, *DependencyNode, error) {
	cache := NewCache()
	rootNode, rootNodeErr := convertFromConfigNode(rootPath, depsPath, conf, cache, mode)
	return cache, rootNode, rootNodeErr
}

func SetupDependencies(filename string, mode Mode, forceClean bool) (*DependencyInfo, error) {
	conf, confErr := ReadConfigFromFilename(filename)
	if confErr != nil {
		return nil, confErr
	}
	packageRootPath := path.Dir(filename)
	rootPath := path.Dir(packageRootPath)
	depsPath := filepath.Join(packageRootPath, "deps/")
	if mode != Clone || forceClean {
		BackupDeps(depsPath)
	}
	os.Mkdir(depsPath, 0700)

	cache, rootNode, rootNodeErr := calculateTotalDependencies(rootPath, depsPath, conf, mode)
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

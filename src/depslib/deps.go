/*---------------------------------------------------------------------------------------------
 *  Copyright (c) Peter Bjorklund. All rights reserved.
 *  Licensed under the MIT License. See LICENSE in the project root for license information.
 *--------------------------------------------------------------------------------------------*/

package depslib

import (
	"fmt"
	"log"
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
	ReadLocal
)

func symlinkRepo(rootPath string, depsPath string, repoName string) error {
	shortName := RepoNameToShortName(repoName)
	packageDir := path.Join(rootPath, shortName+"/")
	targetNameInDeps := path.Join(depsPath, shortName)
	_, statErr := os.Stat(targetNameInDeps)
	if statErr == nil {
		return fmt.Errorf("there is already something at target '%v', can not create link", targetNameInDeps)
	}
	fmt.Printf("symlink '%v' to '%v'\n", packageDir, targetNameInDeps)
	makeErr := MakeSymlink(packageDir, targetNameInDeps)
	if makeErr != nil {
		return makeErr
	}
	return nil
	//return symlinkSrcInclude(packageDir, depsPath, shortName)
}

func wgetRepo(rootPath string, depsPath string, repoName string) error {
	downloadURLString := fmt.Sprintf("https://%vgithub.com/%v/archive/main.zip", gitRepoPrefix(), repoName)
	fmt.Printf("downloading from '%v'\n", downloadURLString)
	downloadURL, parseErr := url.Parse(downloadURLString)
	if parseErr != nil {
		return parseErr
	}

	downloadErr := HTTPGet(downloadURL, "temp.zip")
	if downloadErr != nil {
		return downloadErr
	}

	shortName := RepoNameToShortName(repoName)
	lastName := strings.Split(shortName, "/")[1]
	targetDirectory := path.Join(depsPath, shortName)
	zipPrefix := fmt.Sprintf("%v-main/", lastName)

	unzipErr := unzipFile("temp.zip", targetDirectory, zipPrefix)
	if unzipErr != nil {
		log.Printf("unzipErr:%v", unzipErr)
		return unzipErr
	}
	return nil
}

func gitClone(depsPath string, repoName string, shortName string) error {
	downloadURLString := fmt.Sprintf("https://%vgithub.com/%v.git", gitRepoPrefix(), repoName)
	downloadURL, parseErr := url.Parse(downloadURLString)
	if parseErr != nil {
		return parseErr
	}

	fmt.Printf("git clone from '%v' to %v\n", downloadURL, shortName)

	cmd := exec.Command("git", "clone", downloadURL.String(), shortName)

	cmd.Dir = depsPath

	cmd.Start()

	cmd.Wait()

	return nil
}

func gitPull(targetDirectory string, repoName string) error {
	fmt.Printf("git pull %v %v\n", repoName, targetDirectory)
	cmd := exec.Command("git", "pull")

	cmd.Dir = targetDirectory

	cmd.Start()

	cmd.Wait()

	return nil
}

func gitRepoPrefix() string {
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		return ""
	}

	fmt.Printf("found secret GITHUB_TOKEN\n")

	return fmt.Sprintf("%v@", token)
}

func directoryExists(directory string) bool {
	stat, checkDirectoryErr := os.Lstat(directory)

	return checkDirectoryErr == nil && stat.IsDir()
}

func cloneOrPullRepo(targetDirectory string, depsPath string, repoName string, shortName string) error {
	checkDirectory := path.Join(targetDirectory, ".git")
	if directoryExists(checkDirectory) {
		return gitPull(targetDirectory, repoName)
	}
	return gitClone(depsPath, repoName, shortName)

}

func copyDependency(rootPath string, depsPath string, repoName string, mode Mode) error {
	shortName := RepoNameToShortName(repoName)
	targetDirectory := path.Join(depsPath, shortName)
	fmt.Printf("copy from '%v' to '%v'\n", shortName, targetDirectory)

	if mode != Symlink {
		os.MkdirAll(targetDirectory, 0755)
	} else {
		os.MkdirAll(path.Dir(targetDirectory), 0755)
	}
	switch mode {
	case Symlink:
		return symlinkRepo(rootPath, depsPath, repoName)
	case Clone:
		return cloneOrPullRepo(targetDirectory, depsPath, repoName, shortName)
	case Wget:
		return wgetRepo(rootPath, depsPath, repoName)
	default:
		return fmt.Errorf("unknown mode")
	}
}

func copyOrGetConfigDirectory(rootPath string, depsPath string, repoName string, mode Mode) (string, error) {
	switch mode {
	case ReadLocal:
		shortName := RepoNameToShortName(repoName)
		packageDir := path.Join(rootPath, shortName+"/")
		return packageDir, nil
	default:
		directoryName := RepoNameToShortName(repoName)
		packageDirectory := path.Join(depsPath, directoryName)
		if err := copyDependency(rootPath, depsPath, repoName, mode); err != nil {
			return "", err
		}
		return packageDirectory, nil
	}
}

func establishPackageAndReadConfig(rootPath string, depsPath string, packageName string, mode Mode) (*Config, error) {
	configDirectory, copyErr := copyOrGetConfigDirectory(rootPath, depsPath, packageName, mode)
	if copyErr != nil {
		return nil, copyErr
	}

	conf, confErr := ReadConfigFromDirectory(configDirectory)
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
	Nodes map[string]*DependencyNode
}

func NewCache() *Cache {
	return &Cache{Nodes: make(map[string]*DependencyNode)}
}

func (c *Cache) FindNode(name string) *DependencyNode {
	return c.Nodes[name]
}
func (c *Cache) AddNode(name string, node *DependencyNode) {
	c.Nodes[name] = node
}

func handleNode(rootPath string, depsPath string, node *DependencyNode, cache *Cache, depName string, mode Mode, useDevelopmentDependencies bool) (*DependencyNode, error) {
	foundNode := cache.FindNode(depName)
	if foundNode == nil {
		depConf, confErr := establishPackageAndReadConfig(rootPath, depsPath, depName, mode)
		if confErr != nil {
			return nil, confErr
		}

		var convertErr error
		foundNode, convertErr = convertFromConfigNode(rootPath, depsPath, depConf, cache, mode, useDevelopmentDependencies)
		if convertErr != nil {
			return nil, convertErr
		}
	} else {

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

func convertFromConfigNode(rootPath string, depsPath string, conf *Config, cache *Cache, mode Mode, useDevelopmentDependencies bool) (*DependencyNode, error) {
	artifactType := ToArtifactType(conf.ArtifactType)
	node := &DependencyNode{name: conf.Name, version: semver.MustParse(conf.Version), artifactType: artifactType}
	cache.AddNode(conf.Name, node)
	for _, dep := range conf.Dependencies {
		foundNode, handleErr := handleNode(rootPath, depsPath, node, cache, dep.Name, mode, useDevelopmentDependencies)
		if handleErr != nil {
			return nil, handleErr
		}
		node.AddDependency(foundNode)
	}
	if useDevelopmentDependencies {
		for _, dep := range conf.Development {
			_, handleErr := handleNode(rootPath, depsPath, node, cache, dep.Name, mode, useDevelopmentDependencies)
			if handleErr != nil {
				return nil, handleErr
			}
			//node.AddDevelopment(foundNode)
		}
	}

	return node, nil
}

func CalculateTotalDependencies(rootPath string, depsPath string, conf *Config, mode Mode, useDevelopmentDependencies bool) (*Cache, *DependencyNode, error) {
	cache := NewCache()
	rootNode, rootNodeErr := convertFromConfigNode(rootPath, depsPath, conf, cache, mode, useDevelopmentDependencies)
	return cache, rootNode, rootNodeErr
}

func isInList(dependencies []*DependencyNode, dependencyToCheck *DependencyNode) bool {
	for _, x := range dependencies {
		if x == dependencyToCheck {
			return true
		}
	}

	return false
}

func whoDependsOnThisExcept(dependencies []*DependencyNode, dependencyToCheck *DependencyNode) []*DependencyNode {
	var foundDependencies []*DependencyNode

	for _, x := range dependencies {
		if isInList(x.dependencies, dependencyToCheck) {
			foundDependencies = append(foundDependencies, x)
		}
	}

	return foundDependencies
}

func SetupDependencies(filename string, mode Mode, forceClean bool, useDevelopmentDependencies bool, localPackageRoot string, depsTargetPathOverride string) (*DependencyInfo, error) {
	conf, confErr := ReadConfigFromFilename(filename)
	if confErr != nil {
		return nil, confErr
	}

	var packageRootPath string

	var rootPath string

	if localPackageRoot == "" {
		packageRootPath = path.Dir(filename)
		rootPath = path.Dir(path.Dir(packageRootPath))
	} else {
		rootPath = localPackageRoot
	}

	depsPath := filepath.Join(path.Dir(filename), "deps/")
	if depsTargetPathOverride != "" {
		depsPath = depsTargetPathOverride
	}

	if mode != ReadLocal {
		if mode != Clone || forceClean {
			if err := BackupDeps(depsPath); err != nil {
				return nil, err
			}
		}
		os.Mkdir(depsPath, 0755)
	}

	cache, rootNode, rootNodeErr := CalculateTotalDependencies(rootPath, depsPath, conf, mode, useDevelopmentDependencies)
	if rootNodeErr != nil {
		return nil, rootNodeErr
	}
	var rootNodes []*DependencyNode
	for _, node := range cache.Nodes {
		if node.name == rootNode.name {
			continue
		}
		rootNodes = append(rootNodes, node)
	}

	for _, nodeToCheck := range cache.Nodes {
		for _, localNode := range nodeToCheck.dependencies {
			allDependencies := whoDependsOnThisExcept(nodeToCheck.dependencies, localNode)
			if len(allDependencies) > 0 {
				log.Printf("redundant: '%v' included '%v', but it is already required by '%v'", nodeToCheck, localNode, allDependencies)
			}
		}
	}

	info := &DependencyInfo{RootPath: rootPath, PackageRootPath: packageRootPath, RootNode: rootNode, RootNodes: rootNodes}

	return info, nil
}

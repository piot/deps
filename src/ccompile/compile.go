package ccompile

import (
	"os"
	"path/filepath"

	"github.com/piot/deps/src/depsbuild"
	"github.com/piot/deps/src/depslib"
	"github.com/piot/log-go/src/clog"
)

func directoryExists(directory string) bool {
	stat, checkDirectoryErr := os.Lstat(directory)
	return checkDirectoryErr == nil && stat.IsDir()
}

func fileExists(directory string) bool {
	stat, checkDirectoryErr := os.Lstat(directory)
	return checkDirectoryErr == nil && !stat.IsDir()
}

func libRecursive(searchDir string) ([]string, error) {
	fileList := []string{}
	err := filepath.Walk(searchDir, func(path string, f os.FileInfo, err error) error {
		stat, statErr := os.Lstat(path)
		if statErr != nil {
			return statErr
		}
		if stat.IsDir() {
			cfiles, cfilesErr := filepath.Glob(filepath.Join(path, "*.c"))
			if cfilesErr == nil && len(cfiles) > 0 {
				fileList = append(fileList, path)
			}
		}
		return nil
	})

	return fileList, err
}

func Build(info *depslib.DependencyInfo, log *clog.Log) error {
	depsPath := filepath.Join(info.PackageRootPath, "deps/")
	var sourceLibs []string
	for _, node := range info.RootNodes {
		libPath := filepath.Join(depsPath, node.ShortName(), "src/lib/")
		if directoryExists(libPath) {
			allDirs, recursiveErr := libRecursive(libPath)
			if recursiveErr != nil {
				return recursiveErr
			}
			sourceLibs = append(sourceLibs, allDirs...)
		}
	}
	ownSrcLib := filepath.Join(info.PackageRootPath, "src/lib")
	if directoryExists(ownSrcLib) {
		allOwnSrcLib, allOwnSrcLibErr := libRecursive(ownSrcLib)
		if allOwnSrcLibErr != nil {
			return allOwnSrcLibErr
		}
		sourceLibs = append(sourceLibs, allOwnSrcLib...)
	}

	ownSrcExample := filepath.Join(info.PackageRootPath, "src/examples")
	if directoryExists(ownSrcExample) {
		sourceLibs = append(sourceLibs, ownSrcExample)
	}
	artifactType := info.RootNode.ArtifactType()
	linkFlags := []string{"-lm"}
	localMain := "main.c"
	if fileExists(localMain) {
		artifactType = depslib.Application
		sourceLibs = append(sourceLibs, ".")
		operatingSystem := depsbuild.DetectOS()
		if operatingSystem == depsbuild.MacOS || operatingSystem == depsbuild.Linux {
			log.Debug("adding SDL main")
			sourceLibs = append(sourceLibs, filepath.Join(depsPath, "breathe/src/platform/sdl/"))
			sourceLibs = append(sourceLibs, filepath.Join(depsPath, "burst/src/platform/posix/"))
			linkFlags = append(linkFlags, "-lSDL2")
			linkFlags = append(linkFlags, "-framework OpenGL")

		}
	}

	if artifactType == depslib.Library {
		linkFlags = append(linkFlags, "-shared")
		linkFlags = append(linkFlags, "-fPIC")
	}

	var includePaths []string

	includePaths = append(includePaths, filepath.Join(depsPath, "include"))
	includePaths = append(includePaths, filepath.Join(info.PackageRootPath, "src/include"))

	var defines []string
	defines = append(defines, "_POSIX_C_SOURCE=200112L")
	defines = append(defines, "CONFIGURATION_DEBUG")

	flags := []string{"-g", "--std=c11",
		"-Wall", "-Weverything",
		// "-pedantic", "-Werror",
		"-Wno-sign-conversion", "-Wno-conversion", "-Wno-unused-parameter",
		"-Wno-cast-align",
		"-Wno-padded", "-Wno-cast-qual",
		"-Wno-gnu-folding-constant", "-Wno-unused-macros"}
	operatingSytem := depsbuild.DetectOS()
	switch operatingSytem {
	case depsbuild.MacOS:
		flags = append(flags, "-Wno-extra-semi")
	default:
		flags = append(flags, "-Wno-extra-semi-stmt")
	}
	return depsbuild.Build(flags, sourceLibs, includePaths, defines, linkFlags, log)
}

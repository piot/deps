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

func OSName(os depsbuild.OperatingSystem) string {
	switch os {
	case depsbuild.MacOS:
		return "macos"
	case depsbuild.Linux:
		return "posix"
	case depsbuild.Windows:
		return "windows"
	}

	return ""
}

func platformSpecificPath(depsPath string, node *depslib.DependencyNode, platformName string) string {
	return filepath.Join(depsPath, node.ShortName(), "src/platform/", platformName)
}

func existsPlatformSpecificPath(depsPath string, node *depslib.DependencyNode, platformName string) (string, bool) {
	platformPath := platformSpecificPath(depsPath, node, platformName)
	return platformPath, directoryExists(platformPath)
}

func platformSpecific(depsPath string, node *depslib.DependencyNode, fallbackPlatformName string) string {
	operatingSystem := depsbuild.DetectOS()

	calculatedPath, doesExist := existsPlatformSpecificPath(depsPath, node, OSName(operatingSystem))
	if doesExist {
		return calculatedPath
	}

	return platformSpecificPath(depsPath, node, fallbackPlatformName)
}

func sourceArrayContains(sources []string, query string) bool {
	for _, s := range sources {
		if query == s {
			return true
		}
	}
	return false
}

func findPlatformSpecific(depsPath string, node *depslib.DependencyNode) string {
	fallbackPlatformName := "posix"

	return platformSpecific(depsPath, node, fallbackPlatformName)
}

func Build(info *depslib.DependencyInfo, artifactTypeOverride depslib.ArtifactType, log *clog.Log) ([]string, error) {
	depsPath := filepath.Join(info.PackageRootPath, "deps/")

	var sourceLibs []string

	for _, node := range info.RootNodes {
		libPath := filepath.Join(depsPath, node.ShortName(), "src/lib/")
		if directoryExists(libPath) {
			allDirs, recursiveErr := libRecursive(libPath)
			if recursiveErr != nil {
				return nil, recursiveErr
			}

			sourceLibs = append(sourceLibs, allDirs...)
		}

		sdlSpecific := filepath.Join(depsPath, node.ShortName(), "src/platform/sdl")
		//nolint: nestif
		if directoryExists(sdlSpecific) {
			allDirs, recursiveErr := libRecursive(sdlSpecific)
			if recursiveErr != nil {
				return nil, recursiveErr
			}

			sourceLibs = append(sourceLibs, allDirs...)

			sdlCommon := filepath.Join(depsPath, node.ShortName(), "src/platform/sdl_common")
			if directoryExists(sdlCommon) {
				allDirs, recursiveErr := libRecursive(sdlCommon)
				if recursiveErr != nil {
					return nil, recursiveErr
				}

				sourceLibs = append(sourceLibs, allDirs...)
			}
		} else {
			platformSpecific := findPlatformSpecific(depsPath, node)
			if directoryExists(platformSpecific) {
				allDirs, recursiveErr := libRecursive(platformSpecific)
				if recursiveErr != nil {
					return nil, recursiveErr
				}
				sourceLibs = append(sourceLibs, allDirs...)
			}
		}
	}

	ownSrcLib := filepath.Join(info.PackageRootPath, "src/lib")
	if directoryExists(ownSrcLib) {
		allOwnSrcLib, allOwnSrcLibErr := libRecursive(ownSrcLib)
		if allOwnSrcLibErr != nil {
			return nil, allOwnSrcLibErr
		}

		sourceLibs = append(sourceLibs, allOwnSrcLib...)
	}

	useSDL := false

	artifactType := info.RootNode.ArtifactType()
	linkFlags := []string{"-lm"}

	localMain := "main.c"
	//nolint: nestif
	if fileExists(localMain) {
		if artifactType == depslib.Library {
			artifactType = depslib.Application
		}

		if artifactTypeOverride != depslib.Inherit {
			artifactType = artifactTypeOverride
		}

		thisDirectory, _ := filepath.Abs(".")
		if !sourceArrayContains(sourceLibs, thisDirectory) {
			sourceLibs = append(sourceLibs, thisDirectory)
		}

		operatingSystem := depsbuild.DetectOS()
		if operatingSystem == depsbuild.MacOS || operatingSystem == depsbuild.Linux {
			if artifactType == depslib.ConsoleApplication {
				log.Info("adding console main")
			} else {
				log.Info("adding SDL main")
				useSDL = true
				linkFlags = append(linkFlags, "-lSDL2")
				if operatingSystem == depsbuild.MacOS {
					linkFlags = append(linkFlags, "-framework OpenGL")
				} else {
					linkFlags = append(linkFlags, "-lGL")
				}
			}

		}
	}

	if artifactType == depslib.Library {
		linkFlags = append(linkFlags, "-shared")
		linkFlags = append(linkFlags, "-fPIC")
	}

	var includePaths []string

	includePaths = append(includePaths, filepath.Join(depsPath, "include"))
	includePaths = append(includePaths, filepath.Join(info.PackageRootPath, "src/include"))

	if useSDL {
		includePaths = append(includePaths, "/usr/include/SDL2/")
	}

	var defines []string
	defines = append(defines, "_POSIX_C_SOURCE=200112L")
	defines = append(defines, "CONFIGURATION_DEBUG")
	defines = append(defines, "TYRAN_CONFIGURATION_DEBUG")

	flags := []string{"-g", "-O0", "--std=c11",
		"-Wall", "-Weverything",
		"-Wno-disabled-macro-expansion", "-Wno-reserved-id-macro", "-Wno-documentation", "-Wno-comma",
		"-Wno-double-promotion", "-Wno-c++-compat", "-Wno-covered-switch-default",
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

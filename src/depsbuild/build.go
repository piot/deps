/*---------------------------------------------------------------------------------------------
 *  Copyright (c) Peter Bjorklund. All rights reserved.
 *  Licensed under the MIT License. See LICENSE in the project root for license information.
 *--------------------------------------------------------------------------------------------*/

package depsbuild

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

type OperatingSystem uint

const (
	Linux OperatingSystem = iota
	MacOS
	Windows
	Unknown
)

func DetectOS() OperatingSystem {
	os := runtime.GOOS

	if strings.Contains(os, "windows") {
		return Windows
	} else if strings.Contains(os, "linux") {
		return Linux
	} else if strings.Contains(os, "darwin") {
		return MacOS
	}

	return Unknown
}

func OSSuffix() string {
	switch DetectOS() {
	case Windows:
		return "WINDOWS"
	case MacOS:
		return "MAC_OS_X"
	case Linux:
		return "LINUX"
	}
	return "UNKNOWN"
}

func OSDefine() string {
	return "TORNADO_OS_" + OSSuffix()
}

func Prefix(sources []string, prefix string) string {
	result := ""
	for index, path := range sources {
		if index > 0 {
			result += " "
		}
		result += prefix + path
	}
	return result
}

func PrefixFile(sources []string, prefix string, wd string) string {
	result := ""
	for index, path := range sources {
		newPath, newPathErr := filepath.Rel(wd, path)
		if newPathErr != nil {
			newPath = path
		}

		if index > 0 {
			result += " "
		}
		result += prefix + newPath
	}
	return result
}

func SuffixExtension(sources []string, suffix string, wd string) string {
	result := ""
	for index, path := range sources {
		if index > 0 {
			result += " "
		}
		newPath, newPathErr := filepath.Rel(wd, path)
		if newPathErr != nil {
			newPath = path
		}
		completePath := filepath.Join(newPath, suffix)
		result += completePath
	}
	return result
}

func Build(flags []string, sources []string, includes []string, defines []string, linkFlags []string) ([]string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	outputFilename := "./a.out"
	_ = os.Remove(outputFilename)

	compileSources := SuffixExtension(sources, "*.c", dir)
	allDefines := append(defines, OSDefine())
	defineString := Prefix(allDefines, "-D ")
	includeString := PrefixFile(includes, "-I ", dir)
	flagString := strings.Join(flags, " ")
	linkFlagString := strings.Join(linkFlags, " ")
	executeErr := Execute("clang", flagString, compileSources, defineString, includeString, linkFlagString)
	if executeErr != nil {
		return nil, executeErr
	}

	return []string{outputFilename}, nil
}

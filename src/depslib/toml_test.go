/*---------------------------------------------------------------------------------------------
 *  Copyright (c) Peter Bjorklund. All rights reserved.
 *  Licensed under the MIT License. See LICENSE in the project root for license information.
 *--------------------------------------------------------------------------------------------*/

package depslib

import (
	"testing"
)

func TestToml(t *testing.T) {
	conf, err := ReadConfigFromDirectory("../../test/first/")
	if err != nil {
		t.Fatal(err)
	}

	count := len(conf.Dependencies)
	if count != 2 {
		t.Fatalf("wrong package count %d", count)
	}
	packageNameToTest := conf.Dependencies[1].Name

	if packageNameToTest != "piot/tiny-clib" {
		t.Errorf("wrong package name %v", packageNameToTest)
	}
}

/*---------------------------------------------------------------------------------------------
 *  Copyright (c) Peter Bjorklund. All rights reserved.
 *  Licensed under the MIT License. See LICENSE in the project root for license information.
 *--------------------------------------------------------------------------------------------*/

package depsbuild

import (
	"os"
	"os/exec"
	"strings"
)

func Execute(executable string, cmdStrings ...string) error {
	debugString := executable + " " + strings.Join(cmdStrings, " ")
	cmd := exec.Command("bash", "-c", debugString)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	return cmd.Run()
}

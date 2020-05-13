package depsbuild

import (
	"os"
	"os/exec"
	"strings"

	"github.com/piot/log-go/src/clog"
)

func Execute(log *clog.Log, executable string, cmdStrings ...string) error {
	debugString := executable + " " + strings.Join(cmdStrings, " ")
	log.Info("executing", clog.String("cmd", debugString))
	cmd := exec.Command("bash", "-c", debugString)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	return cmd.Run()
}

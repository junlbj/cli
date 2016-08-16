package v2

import (
	"os"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/commands"
	"code.cloudfoundry.org/cli/commands/flags"
)

type EnableSSHCommand struct {
	RequiredArgs flags.AppName `positional-args:"yes"`
	usage        interface{}   `usage:"CF_NAME enable-ssh APP_NAME"`
}

func (_ EnableSSHCommand) Setup(config commands.Config) error {
	return nil
}

func (_ EnableSSHCommand) Execute(args []string) error {
	cmd.Main(os.Getenv("CF_TRACE"), os.Args)
	return nil
}

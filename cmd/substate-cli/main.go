package main

import (
	"fmt"
	"os"

	"github.com/Fantom-foundation/substate-cli/cmd/substate-cli/db"
	"github.com/Fantom-foundation/substate-cli/cmd/substate-cli/replay"
	"github.com/Fantom-foundation/go-opera/flags"
	"github.com/ethereum/go-ethereum/substate"
	cli "gopkg.in/urfave/cli.v1"
)

var (
	dbCommand = cli.Command{
		Name:        "db",
		Usage:       "A set of commands on substate DB",
		Description: "",
		Subcommands: []cli.Command{
			db.CloneCommand,
			db.CompactCommand,
		},
	}
)

var (
	gitCommit = "" // Git SHA1 commit hash of the release (set via linker flags)
	gitDate   = ""

	app = flags.NewApp(gitCommit, gitDate, "Fantom substate command line interface")
)

func init() {
	app.Flags = []cli.Flag{}
	app.Commands = []cli.Command{
		replay.ReplayCommand,
		replay.GetStorageUpdateSizeCommand,
		replay.GetCodeCommand,
		replay.GetCodeSizeCommand,
		replay.SubstateDumpCommand,
		dbCommand,
	}
	cli.CommandHelpTemplate = flags.CommandHelpTemplate 
}

func main() {
	substate.RecordReplay = true
	if err := app.Run(os.Args); err != nil {
		code := 1
		fmt.Fprintln(os.Stderr, err)
		os.Exit(code)
	}
}

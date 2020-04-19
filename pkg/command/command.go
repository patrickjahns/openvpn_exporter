package command

import (
	"github.com/patrickjahns/openvpn_exporter/pkg/version"
	"os"

	"github.com/urfave/cli/v2"
)

// Run parses the command line arguments and executes the program.
func Run() error {

	app := &cli.App{
		Name:    "openvpn_exporter",
		Version: version.String + " (" + version.Revision + " // " + version.Date + ")",
		Usage:   "OpenVPN exporter",
		Authors: []*cli.Author{
			{
				Name:  "Patrick Jahns",
				Email: "github@patrickjahns.de",
			},
		},
	}

	cli.HelpFlag = &cli.BoolFlag{
		Name:    "help",
		Aliases: []string{"h"},
		Usage:   "Show help",
	}

	cli.VersionFlag = &cli.BoolFlag{
		Name:    "version",
		Aliases: []string{"v"},
		Usage:   "Prints the current version",
	}

	return app.Run(os.Args)
}

package main

import (
	"context"
	"log"
	"os"

	cli "github.com/urfave/cli/v3"

	hostname_cname "github.com/nicksantamaria/uri-rewriter/internal/cmd/hostname-cname"
	"github.com/nicksantamaria/uri-rewriter/internal/cmd/version"
)

var (
	buildVersion = "dev"
	buildCommit  = "none"
	buildDate    = "unknown"
)

func main() {
	ctx := context.Background()

	// Set version information in the version package
	ctx = context.WithValue(ctx, "buildVersion", buildVersion)
	ctx = context.WithValue(ctx, "buildCommit", buildCommit)
	ctx = context.WithValue(ctx, "buildDate", buildDate)

	app := &cli.Command{
		Name:  "uri-rewriter",
		Usage: "Rewrite URI hostnames using CNAME records",
		Commands: []*cli.Command{
			{
				Name:  "hostname-cname",
				Usage: "Replace the URI's hostname with its CNAME at a given level",
				Flags: []cli.Flag{
					&cli.IntFlag{
						Name:    "cname-level",
						Aliases: []string{"l"},
						Value:   1,
						Usage:   "CNAME level to use (1 = first CNAME in the chain)",
					},
				},
				ArgsUsage: "<uri>",
				Action:    hostname_cname.RunHostnameCNAME,
			},
			{
				Name:   "version",
				Usage:  "Display the application version information",
				Action: version.DisplayVersion,
			},
		},
	}

	if err := app.Run(ctx, os.Args); err != nil {
		log.Fatal(err)
	}
}

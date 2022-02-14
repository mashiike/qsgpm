package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sort"
	"strings"
	"syscall"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/fatih/color"
	"github.com/fujiwara/logutils"
	"github.com/mashiike/qsgpm"
	"github.com/urfave/cli/v2"
)

var Version string = "current"
var filter = &logutils.LevelFilter{
	Levels: []logutils.LogLevel{"debug", "info", "notice", "warn", "error"},
	ModifierFuncs: []logutils.ModifierFunc{
		logutils.Color(color.FgHiBlack),
		logutils.Color(color.FgWhite),
		logutils.Color(color.FgHiBlue),
		logutils.Color(color.FgYellow),
		logutils.Color(color.FgRed, color.Bold),
	},
	Writer: os.Stderr,
}

func main() {
	cliApp := &cli.App{
		Name:      "qsgpm",
		Usage:     "A commandline tool for management of QuickSight Group and CustomPermission",
		UsageText: "qsgpm -config <config file>",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "config",
				Aliases: []string{"c"},
				Usage:   "config file path",
				EnvVars: []string{"CONFIG", "QSGPM_CONFIG"},
			},
			&cli.StringFlag{
				Name:    "log-level",
				Aliases: []string{"l"},
				Usage:   "output log level (debug|info|notice|warn|error)",
				Value:   "info",
				EnvVars: []string{"QSGPM_LOG_LEVEL"},
			},
			&cli.BoolFlag{
				Name:    "dry-run",
				EnvVars: []string{"QSGPM_DRY_RUN"},
			},
		},
		Action: run,
	}
	sort.Sort(cli.FlagsByName(cliApp.Flags))
	cliApp.Version = Version
	cliApp.EnableBashCompletion = true
	cliApp.Before = func(c *cli.Context) error {
		filter.MinLevel = logutils.LogLevel(c.String("log-level"))
		log.SetOutput(filter)
		log.Println("[debug] set log level:", c.String("log-level"))
		return nil
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT, syscall.SIGHUP)
	defer cancel()
	if err := cliApp.RunContext(ctx, os.Args); err != nil {
		log.Printf("[error] %s", err)
	}
}

func isLambda() bool {
	return strings.HasPrefix(os.Getenv("AWS_EXECUTION_ENV"), "AWS_Lambda") ||
		os.Getenv("AWS_LAMBDA_RUNTIME_API") != ""
}

func run(c *cli.Context) error {

	cfg := qsgpm.NewDefaultConfig()
	if err := cfg.Load(c.String("config")); err != nil {
		return err
	}
	if err := cfg.ValidateVersion(Version); err != nil {
		return err
	}
	app, err := qsgpm.New(c.Context, cfg)
	if err != nil {
		return err
	}
	if isLambda() {
		lambda.Start(func(ctx context.Context) error {
			return app.Run(ctx, qsgpm.RunOption{
				DryRun: false,
			})
		})
		return nil
	}
	return app.Run(c.Context, qsgpm.RunOption{
		DryRun: c.Bool("dry-run"),
	})
}

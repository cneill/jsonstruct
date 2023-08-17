package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/cneill/jsonstruct"

	"github.com/urfave/cli/v2"
)

var log *slog.Logger

func run() error {
	logHandler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo})
	log = slog.New(logHandler)

	app := &cli.App{
		Name:        "jsonstruct",
		Action:      genStructs,
		ArgsUsage:   "[file]...",
		Usage:       "generate Go structs for JSON values",
		Description: `You can either pass in files as args or JSON in STDIN. Results are printed to STDOUT.`,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "name",
				Aliases: []string{"n"},
				Usage:   "override the default name derived from filename",
			},
			&cli.BoolFlag{
				Name:    "value-comments",
				Aliases: []string{"c"},
				Usage:   "add a comment to struct fields with the example value(s)",
			},
			&cli.BoolFlag{
				Name:    "sort-fields",
				Aliases: []string{"s"},
				Usage:   "sort the fields in alphabetical order; default behavior is to mirror input",
			},
			&cli.BoolFlag{
				Name:    "inline",
				Aliases: []string{"i"},
				Usage:   "use inline structs instead of creating different types for each object",
			},
			&cli.BoolFlag{
				Name:  "debug",
				Value: false,
				Usage: "enable debug logs",
			},
		},
		Before: func(ctx *cli.Context) error {
			// set debug logging
			if ctx.Bool("debug") {
				logHandler = slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug})
				log = slog.New(logHandler)
			}

			return nil
		},
	}

	return app.Run(os.Args)
}

func isStdin() bool {
	stat, _ := os.Stdin.Stat()

	return (stat.Mode() & os.ModeCharDevice) == 0
}

func genStructs(ctx *cli.Context) error {
	inputs := []*os.File{}

	formatter, err := jsonstruct.NewFormatter(&jsonstruct.FormatterOptions{
		SortFields:    ctx.Bool("sort-fields"),
		ValueComments: ctx.Bool("value-comments"),
	})
	if err != nil {
		return fmt.Errorf("failed to set up formatter: %w", err)
	}

	if isStdin() {
		inputs = append(inputs, os.Stdin)
	} else {
		for _, fileName := range ctx.Args().Slice() {
			file, err := os.Open(fileName)
			if err != nil {
				return fmt.Errorf("failed to open file %q: %w", fileName, err)
			}

			inputs = append(inputs, file)
		}
	}

	for _, input := range inputs {
		defer func() {
			input.Close()
		}()

		p := jsonstruct.NewParser(input, log)

		results, err := p.Start()
		if err != nil {
			return fmt.Errorf("failed to parse input %q: %w", input.Name(), err)
		}

		fmt.Println(input.Name())

		if err := formatter.Format(results...); err != nil {
			return fmt.Errorf("failed to format struct(s) from %q: %w", input.Name(), err)
		}
	}

	return nil
}

func main() {
	if err := run(); err != nil {
		log.Error("failed to execute", "err", err)
	}
}

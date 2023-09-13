package main

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/cneill/jsonstruct"

	"github.com/urfave/cli/v2"
)

//nolint:gochecknoglobals // c'mon, it's the logger
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
				Name:    "inline-structs",
				Aliases: []string{"i"},
				Usage:   "use inline structs instead of creating different types for each object",
			},
			&cli.BoolFlag{
				Name:    "print-filenames",
				Aliases: []string{"f"},
				Usage:   "print the filename above the structs defined within",
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

	if err := app.Run(os.Args); err != nil {
		return fmt.Errorf("ERROR: %w", err)
	}

	return nil
}

func isStdin() bool {
	stat, _ := os.Stdin.Stat()

	return (stat.Mode() & os.ModeCharDevice) == 0
}

func genStructs(ctx *cli.Context) error {
	inputs, err := getInputs(ctx)
	if err != nil {
		return err
	}

	formatter, err := jsonstruct.NewFormatter(&jsonstruct.FormatterOptions{
		SortFields:    ctx.Bool("sort-fields"),
		ValueComments: ctx.Bool("value-comments"),
		InlineStructs: ctx.Bool("inline-structs"),
	})
	if err != nil {
		return fmt.Errorf("failed to set up formatter: %w", err)
	}

	for _, input := range inputs {
		defer func() {
			input.Close()

			log.Debug("closed file", "file", input.Name())
		}()

		p := jsonstruct.NewParser(input, log)

		results, err := p.Start()
		if err != nil {
			return fmt.Errorf("failed to parse input %q: %w", input.Name(), err)
		}

		goFileName := jsonstruct.GetFileGoName(input.Name())

		for i := 0; i < len(results); i++ {
			structName := fmt.Sprintf("%s%d", goFileName, i+1)
			results[i].SetName(structName)
		}

		if ctx.Bool("print-filenames") {
			spacer := strings.Repeat("=", len(input.Name()))
			fmt.Printf("// %s\n// %s\n// %s\n", spacer, input.Name(), spacer)
		}

		result, err := formatter.FormatStructs(results...)
		if err != nil {
			return fmt.Errorf("failed to format structs: %w", err)
		}

		fmt.Printf("%s\n", result)
	}

	return nil
}

func getInputs(ctx *cli.Context) ([]*os.File, error) {
	inputs := []*os.File{}

	if isStdin() {
		inputs = append(inputs, os.Stdin)

		log.Debug("got JSON input from stdin")
	} else {
		for _, fileName := range ctx.Args().Slice() {
			file, err := os.Open(fileName)
			if err != nil {
				return nil, fmt.Errorf("failed to open file %q: %w", fileName, err)
			}
			log.Debug("opened file to read JSON structs", "file", fileName)

			inputs = append(inputs, file)
		}
	}

	return inputs, nil
}

func main() {
	if err := run(); err != nil {
		log.Error("failed to execute", "err", err)
	}
}

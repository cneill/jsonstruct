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
var (
	logHandler = slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo})
	log        = slog.New(logHandler)
)

func run() error {
	app := &cli.App{
		Name:        "jsonstruct",
		Action:      genStructs,
		ArgsUsage:   "[FILE]...",
		Usage:       "generate Go structs for JSON values",
		Description: "You can either pass in files as args or JSON in STDIN. Results are printed to STDOUT.",
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
			&cli.StringFlag{
				Name:    "out-file",
				Aliases: []string{"o"},
				Usage:   "write the results to `FILE`",
			},
			&cli.BoolFlag{
				Name:    "debug",
				Aliases: []string{"d"},
				Value:   false,
				Usage:   "enable debug logs",
			},
		},
		Before: setDebug,
		Commands: []*cli.Command{
			httpCommand(),
		},
	}

	if err := app.Run(os.Args); err != nil {
		return fmt.Errorf("ERROR: %w", err)
	}

	return nil
}

func setDebug(ctx *cli.Context) error {
	// set debug logging
	if ctx.Bool("debug") {
		logHandler = slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug})
		log = slog.New(logHandler)
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

	if len(inputs) == 0 {
		cli.ShowAppHelpAndExit(ctx, 1)
	}

	outFile := os.Stdout

	if outputPath := ctx.String("out-file"); outputPath != "" {
		outFile, err = os.Create(outputPath)
		if err != nil {
			return fmt.Errorf("failed to create out-file %q: %w", outputPath, err)
		}

		defer outFile.Close()
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
		jStructs, err := parseInput(input)
		if err != nil {
			return err
		}

		// print out comments with the name of the file where we saw the struct
		if ctx.Bool("print-filenames") {
			spacer := strings.Repeat("=", len(input.Name()))
			fmt.Fprintf(outFile, "// %s\n// %s\n// %s\n", spacer, input.Name(), spacer)
		}

		result, err := formatter.FormatStructs(jStructs...)
		if err != nil {
			return fmt.Errorf("failed to format structs: %w", err)
		}

		fmt.Fprintf(outFile, "%s\n", result)
	}

	return nil
}

func parseInput(input *os.File) (jsonstruct.JSONStructs, error) {
	defer func() {
		input.Close()

		log.Debug("closed input file", "file", input.Name())
	}()

	parser := jsonstruct.NewParser(input, log)

	jStructs, err := parser.Start()
	if err != nil {
		return nil, fmt.Errorf("failed to parse input %q: %w", input.Name(), err)
	}

	// set the names of the top-level structs from our example file based on the file's name
	goFileName := jsonstruct.GetFileGoName(input.Name())

	for i := 0; i < len(jStructs); i++ {
		structName := fmt.Sprintf("%s%d", goFileName, i+1)
		jStructs[i].SetName(structName)
	}

	return jStructs, nil
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

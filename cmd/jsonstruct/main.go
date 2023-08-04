package main

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/cneill/jsonstruct"

	"github.com/urfave/cli/v2"
)

func isStdin() bool {
	stat, _ := os.Stdin.Stat()

	return (stat.Mode() & os.ModeCharDevice) == 0
}

func run() error {
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
		},
	}

	if err := app.Run(os.Args); err != nil {
		return fmt.Errorf("ERROR: %w", err)
	}

	return nil
}

func genStructs(ctx *cli.Context) error {
	jsp := &jsonstruct.Producer{
		SortFields:    ctx.Bool("sort-fields"),
		ValueComments: ctx.Bool("value-comments"),
		Name:          ctx.String("name"),
	}

	results := []*jsonstruct.JSONStruct{}

	if isStdin() {
		contents, err := io.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("error reading from stdin: %w", err)
		}

		result, err := jsp.StructFromBytes("stdin", contents)
		if err != nil {
			return fmt.Errorf("failed to parse stdin: %w", err)
		}

		results = append(results, result)
	} else {
		for _, file := range ctx.Args().Slice() {
			result, err := jsp.StructFromExampleFile(file)
			if err != nil {
				return fmt.Errorf("failed to parse file %q: %w", file, err)
			}

			results = append(results, result)
		}
	}

	for _, result := range results {
		formatted, err := jsonstruct.GoFmt(result)
		if err != nil {
			fmt.Printf("%s\n", result.String())
		} else {
			fmt.Println(formatted)
		}
	}

	return nil
}

func main() {
	if err := run(); err != nil {
		log.Fatalf("%v\n", err)
	}
}

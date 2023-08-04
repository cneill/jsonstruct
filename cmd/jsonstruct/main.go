package main

import (
	"fmt"
	"log"
	"os"

	"github.com/cneill/jsonstruct"

	"github.com/urfave/cli/v2"
)

var packagePrefix = "package temp\n"

func isStdin() bool {
	stat, _ := os.Stdin.Stat()

	return (stat.Mode() & os.ModeCharDevice) == 0
}

func run() error {
	app := &cli.App{
		Name:        "jsonstruct",
		Action:      genStruct,
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

func genStruct(ctx *cli.Context) error {
	jsp := &jsonstruct.Producer{
		SortFields:    ctx.Bool("sort-fields"),
		ValueComments: ctx.Bool("value-comments"),
		Name:          ctx.String("name"),
	}

	results := []*jsonstruct.JSONStruct{}

	if isStdin() {
		js, err := jsp.StructFromStdin()
		if err != nil {
			return err
		}

		results = append(results, js)
	} else {
		for _, file := range ctx.Args().Slice() {
			js, err := jsp.StructFromExampleFile(file)
			if err != nil {
				return err
			}

			results = append(results, js)
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

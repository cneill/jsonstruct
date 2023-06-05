package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/cneill/jsonstruct"
)

var (
	name          string
	sortFields    bool
	valueComments bool
)

func init() {
	flag.StringVar(&name, "name", "", "override the main structs of all passed in files/stdin objects")
	flag.BoolVar(&valueComments, "value-comments", false, "add a comment to struct fields with the example value(s)")
	flag.BoolVar(&sortFields, "sort-fields", true, "sort the fields in alphabetical order")

	flag.Usage = func() {
		fmt.Printf("Usage of %s:\n", os.Args[0])
		fmt.Printf("%s [flags] [file name...]\n\n", os.Args[0])
		fmt.Printf("You can also pass in JSON to stdin if you prefer.")
		fmt.Println("Flags:")
		flag.PrintDefaults()
	}
}

var packagePrefix = "package temp\n"

func goFmt(structs ...*jsonstruct.JSONStruct) (string, error) {
	goFmt, err := exec.LookPath("gofmt")
	if err != nil {
		return "", err
	}

	f, err := ioutil.TempFile("", "jsonstruct-*")
	if err != nil {
		return "", err
	}

	defer f.Close()
	defer os.Remove(f.Name())

	contents := packagePrefix
	for _, js := range structs {
		contents += js.String() + "\n"
	}

	if _, err := f.WriteString(contents); err != nil {
		return "", err
	}

	cmd := exec.Command(goFmt, "-e", f.Name())
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("%s\n", output)
		return "", fmt.Errorf("failed to execute gofmt: %v", err)
	}

	outputStr := strings.TrimSpace(strings.TrimPrefix(string(output), packagePrefix))

	return outputStr, nil
}

func isStdin() bool {
	stat, _ := os.Stdin.Stat()
	return (stat.Mode() & os.ModeCharDevice) == 0
}

func run() error {
	flag.Parse()

	if flag.NArg() < 1 && !isStdin() {
		flag.Usage()

		return fmt.Errorf("must supply one or more file names, or provide something from stdin")
	}

	jsp := &jsonstruct.Producer{
		SortFields:           sortFields,
		VerboseValueComments: valueComments,
		Name:                 name,
	}

	results := []*jsonstruct.JSONStruct{}

	if isStdin() {
		js, err := jsp.StructFromStdin()
		if err != nil {
			return err
		}

		results = append(results, js)
	} else {
		for _, file := range flag.Args() {
			js, err := jsp.StructFromExampleFile(file)
			if err != nil {
				return err
			}

			results = append(results, js)
		}
	}

	for _, result := range results {
		formatted, err := goFmt(result)
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

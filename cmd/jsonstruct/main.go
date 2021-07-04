package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/cneill/jsonstruct"
)

var (
	valueComments bool
)

func init() {
	flag.BoolVar(&valueComments, "value-comments", false, "add a comment to struct fields with the example value(s)")
}

func errh(err error) {
	if err != nil {
		panic(err)
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

	var contents = packagePrefix
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

func main() {
	if len(os.Args) < 2 {
		errh(fmt.Errorf("must supply a file name"))
	}

	flag.Parse()

	jsp := &jsonstruct.Producer{
		VerboseValueComments: valueComments,
	}
	js, err := jsp.StructsFromExampleFile(os.Args[1])
	errh(err)

	formatted, err := goFmt(js)
	if err != nil {
		fmt.Printf("%s\n", js.String())
	} else {
		fmt.Println(formatted)
	}
}

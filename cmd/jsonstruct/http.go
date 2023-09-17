package main

import (
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"strings"

	"github.com/cneill/jsonstruct"
	"github.com/urfave/cli/v2"
)

//go:embed http/static/*
var staticContent embed.FS

//go:embed http/templates
var templatesContent embed.FS

func httpListener(ctx *cli.Context) error {
	staticFS, err := fs.Sub(staticContent, "http/static")
	if err != nil {
		return fmt.Errorf("failed to set up static files: %w", err)
	}

	mux := http.NewServeMux()

	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticFS))))
	mux.HandleFunc("/generate", GenerateHandler)
	mux.HandleFunc("/", IndexHandler)

	listen := fmt.Sprintf("%s:%d", ctx.String("host"), ctx.Int("port"))

	fmt.Printf("Listening on %s...\n", listen)

	if err := http.ListenAndServe(listen, mux); err != nil {
		return fmt.Errorf("listening error: %w", err)
	}

	return nil
}

// GenerateHandler serves the generated content.
func GenerateHandler(w http.ResponseWriter, req *http.Request) {
	t, err := template.New("generate").ParseFS(templatesContent, "http/templates/*.gohtml")
	if err != nil {
		doErr(w, fmt.Errorf("failed to load generate template: %w", err))

		return
	}

	if err := req.ParseForm(); err != nil {
		doErr(w, fmt.Errorf("failed to parse form: %w", err))

		return
	}

	input := req.PostForm.Get("input")
	if input == "" {
		return
	}

	r := strings.NewReader(input)

	parser := jsonstruct.NewParser(r, log)

	jStructs, err := parser.Start()
	if err != nil {
		doErr(w, fmt.Errorf("failed to parse input: %w", err))
		return
	}

	// set the names of the top-level structs from our example file based on the file's name
	name := "WebGenerated"

	for i := 0; i < len(jStructs); i++ {
		jStructs[i].SetName(fmt.Sprintf("%s%d", name, i+1))
	}

	formatter, err := jsonstruct.NewFormatter(&jsonstruct.FormatterOptions{
		SortFields:    req.PostForm.Get("sort_fields") == "on",
		ValueComments: req.PostForm.Get("value_comments") == "on",
		// InlineStructs: ctx.Bool("inline-structs"),
	})
	if err != nil {
		doErr(w, fmt.Errorf("failed to set up formatter: %w", err))
		return
	}

	result, err := formatter.FormatStructs(jStructs...)
	if err != nil {
		doErr(w, fmt.Errorf("failed to format structs: %w", err))
		return
	}

	// fmt.Fprintf(outFile, "%s\n", result)

	data := struct {
		Generated string
	}{result}

	if err := t.Execute(w, data); err != nil {
		doErr(w, fmt.Errorf("failed to execute generate template: %w", err))
		return
	}
}

// IndexHandler serves the main page.
func IndexHandler(w http.ResponseWriter, _ *http.Request) {
	t, err := template.New("index").ParseFS(templatesContent, "http/templates/*.gohtml")
	if err != nil {
		doErr(w, fmt.Errorf("failed to load index template: %w", err))
		return
	}

	if err := t.Execute(w, nil); err != nil {
		doErr(w, fmt.Errorf("failed to execute index template: %w", err))
		return
	}
}

func doErr(w http.ResponseWriter, err error) {
	fmt.Printf("ERROR WITH REQUEST: %v\n", err)
	w.WriteHeader(http.StatusBadRequest)

	if _, err := w.Write([]byte(fmt.Sprintf("%v", err))); err != nil {
		fmt.Printf("error writing error: %v\n", err)
	}
}

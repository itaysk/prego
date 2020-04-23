package main

import (
	"encoding/json"
	"fmt"
	"os"
	"text/template"

	"github.com/itaysk/regogo"
	"github.com/open-policy-agent/opa/rego"
)

// Printer is an interface for printing prego results
type Printer interface {
	// Preamble prints something before results printing begins (one time)
	Preamble()
	// Epilogue prints something after results printing ends (one time)
	Epilogue()
	// Print prints a single resultset
	Print(results rego.ResultSet)
}

type jsonPrinter struct{}

func (p jsonPrinter) Preamble() {}

func (p jsonPrinter) Print(results rego.ResultSet) {
	resBytes, _ := json.Marshal(results)
	fmt.Println(resBytes)
}

func (p jsonPrinter) Epilogue() {}

type regogoPrinter struct {
	query string
}

func (p regogoPrinter) Preamble() {}

func (p regogoPrinter) Print(results rego.ResultSet) {
	for _, result := range results {
		// TODO: modify regogo so it can parse once
		resultBytes, _ := json.Marshal(result)
		regogoResult, _ := regogo.Get(string(resultBytes), p.query)
		fmt.Println(regogoResult.JSON())
	}
}

func (p regogoPrinter) Epilogue() {}

type gotemplatePrinter struct {
	template *template.Template
}

func (p gotemplatePrinter) Preamble() {}

func (p gotemplatePrinter) Print(results rego.ResultSet) {
	for _, result := range results {
		p.template.Execute(os.Stdout, result)
		fmt.Println()
	}
}

func (p gotemplatePrinter) Epilogue() {}

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"text/template"

	"github.com/itaysk/regogo"
)

// Printer is an interface for printing prego results
type Printer interface {
	// Preamble prints something before results printing begins (one time)
	Preamble()
	// Epilogue prints something after results printing ends (one time)
	Epilogue()
	// Print prints a single result
	Print(results interface{})
}

type jsonPrinter struct{}

func (p jsonPrinter) Preamble() {}

func (p jsonPrinter) Print(results interface{}) {
	resBytes, _ := json.Marshal(results)
	fmt.Println(string(resBytes))
}

func (p jsonPrinter) Epilogue() {}

type regogoPrinter struct {
	rg *regogo.Regogo
}

func (p regogoPrinter) Preamble() {}

func (p regogoPrinter) Print(result interface{}) {
	resultBytes, _ := json.Marshal(result)
	regogoResult, _ := p.rg.Get(string(resultBytes))
	fmt.Println(regogoResult.JSON())
}

func (p regogoPrinter) Epilogue() {}

type gotemplatePrinter struct {
	template *template.Template
}

func (p gotemplatePrinter) Preamble() {}

func (p gotemplatePrinter) Print(result interface{}) {
	p.template.Execute(os.Stdout, result)
	fmt.Println()
}

func (p gotemplatePrinter) Epilogue() {}

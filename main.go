package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"text/template"

	"github.com/itaysk/regogo"
	"github.com/open-policy-agent/opa/rego"
	"github.com/open-policy-agent/opa/storage"
	"github.com/open-policy-agent/opa/storage/inmem"
	"github.com/urfave/cli/v2"
)

var printer Printer

func main() {
	app := &cli.App{
		Usage: "pipe into rego. take json lines from stdin and evaluate them with a loaded rego policy",
		Action: func(c *cli.Context) error {
			policyPaths := c.StringSlice("policy")
			dataPaths := c.StringSlice("data")
			query := c.Path("query")
			stateful := c.Bool("stateful")
			outputFormat := c.String("output")

			outputFormatParts := strings.Split(outputFormat, "=")
			switch outputFormatParts[0] {
			case "json":
				printer = jsonPrinter{}
			case "regogo":
				rg, err := regogo.New(outputFormatParts[1])
				if err != nil {
					return fmt.Errorf("invalid regogo: %s", outputFormatParts[1])
				}
				printer = regogoPrinter{
					rg: rg,
				}
			case "gotemplate":
				t, err := template.New("template").Parse(outputFormatParts[1])
				if err != nil {
					return fmt.Errorf("invalid go template: %s", outputFormatParts[1])
				}
				printer = gotemplatePrinter{
					template: t,
				}
			default:
				return fmt.Errorf("unsupported output format: %s", outputFormat)
			}

			r, err := InitRego(policyPaths, dataPaths, query, stateful)
			if err != nil {
				return err
			}

			// TODO: think of a way to support statefull without an extra query (per event).
			// for example, change the user's original query or find a way to look into data without query
			var stateStore storage.Store
			var qstate rego.PreparedEvalQuery
			if stateful {
				rstate, err := InitRego(policyPaths, dataPaths, "nextstate=data.prego.nextstate", stateful)
				if err != nil {
					return err
				}
				stateStore = inmem.NewFromObject(map[string]interface{}{"prego_state": nil})
				rego.Store(stateStore)(r)
				rego.Store(stateStore)(rstate)
				qstate, err = rstate.PrepareForEval(context.TODO())
				if err != nil {
					return fmt.Errorf("error creating evalQuery: %v", err)
				}
			}
			q, err := r.PrepareForEval(context.TODO())
			if err != nil {
				return fmt.Errorf("error creating evalQuery: %v", err)
			}

			scanner := bufio.NewScanner(os.Stdin)
			printer.Preamble()
			for scanner.Scan() {
				t := scanner.Text()
				var input map[string]interface{}
				json.Unmarshal([]byte(t), &input)
				results, err := q.Eval(context.TODO(), rego.EvalInput(input))
				if err != nil {
					return err
				}
				if len(results) > 0 {
					printer.Print(results)
				}
				if stateful {
					resultsstate, err := qstate.Eval(context.TODO(), rego.EvalInput(input))
					if err != nil {
						return err
					}
					if len(resultsstate) > 0 {
						nextstate, ok := resultsstate[0].Bindings["nextstate"]
						if !ok {
							return fmt.Errorf("can't find nextstate in bindings: %+v", resultsstate)
						}
						storage.WriteOne(context.TODO(), stateStore, storage.ReplaceOp, storage.MustParsePath("/prego_state"), nextstate)
					}
				}
			}
			printer.Epilogue()
			return nil
		},
		Flags: []cli.Flag{
			&cli.StringSliceFlag{
				Name:     "policy",
				Required: true,
				Usage:    "file path to rego policy. can be specified multiple times",
			},
			&cli.StringSliceFlag{
				Name:  "data",
				Usage: "file path to additional data. can be specified multiple times",
			},
			&cli.StringFlag{
				Name:  "query",
				Value: "res = data",
				Usage: "query to evaluate",
			},
			&cli.BoolFlag{
				Name:  "stateful",
				Usage: "feed the evaluation result back to the next evaluation. To use, load a policy using the `prego` package, which defines a rule `nextstate`. This rule will be made available to your poilicy under `data.prego_state`",
			},
			&cli.StringFlag{
				Name:  "output",
				Value: "json",
				Usage: "specify the output format: json/regogo/gotemplate",
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func InitRego(policyPaths []string, dataPaths []string, query string, stateful bool) (*rego.Rego, error) {
	var regoBuilders []func(*rego.Rego)
	regoBuilders = append(regoBuilders, rego.Query(query))
	for _, path := range policyPaths {
		policyBytes, err := ioutil.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("policy file not found: %s", path)
		}
		regoBuilders = append(regoBuilders, rego.Module(path, string(policyBytes)))
	}
	for _, path := range dataPaths {
		dataBytes, err := ioutil.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("data file not found: %s", path)
		}
		regoBuilders = append(regoBuilders, rego.Module(path, string(dataBytes)))
	}
	r := rego.New(regoBuilders...)
	return r, nil
}

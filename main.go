package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/open-policy-agent/opa/rego"
	"github.com/open-policy-agent/opa/storage"
	"github.com/open-policy-agent/opa/storage/inmem"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Usage: "pipe into rego. take json lines from stdin and evaluate them with a loaded rego policy",
		Action: func(c *cli.Context) error {
			policyPaths := c.StringSlice("policy")
			dataPaths := c.StringSlice("data")
			query := c.Path("query")
			feedback := c.Bool("feedback")

			r, err := InitRego(policyPaths, dataPaths, query, feedback)
			if err != nil {
				return err
			}

			var stateStore storage.Store
			var qstate rego.PreparedEvalQuery
			if feedback {
				rstate, err := InitRego(policyPaths, dataPaths, "nextstate=data.prego.nextstate", feedback)
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
			for scanner.Scan() {
				t := scanner.Text()
				var input map[string]interface{}
				json.Unmarshal([]byte(t), &input)
				results, err := q.Eval(context.TODO(), rego.EvalInput(input))
				if err != nil {
					return err
				}
				if len(results) > 0 {
					resBytes, _ := json.Marshal(results)
					fmt.Println(string(resBytes))
				}
				if feedback {
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
				Name:  "feedback",
				Usage: "feed the evaluation result back to the next evaluation. To use, load a policy under the `prego` package, which defines a rule called `nextstate`. This `nextstate` will be nade available to your poilicy under `data.prego_state`",
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func InitRego(policyPaths []string, dataPaths []string, query string, feedback bool) (*rego.Rego, error) {
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

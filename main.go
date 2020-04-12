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
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Usage: "pipe into rego. take json lines from stdin and evaluate them with a loaded rego policy",
		Action: func(c *cli.Context) error {
			policyPaths := c.StringSlice("policy")
			dataPaths := c.StringSlice("data")
			query := c.Path("query")
			q, err := InitRego(policyPaths, dataPaths, query)
			if err != nil {
				return err
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
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func InitRego(policyPaths []string, dataPaths []string, query string) (rego.PreparedEvalQuery, error) {
	var regoBuilders []func(*rego.Rego)
	regoBuilders = append(regoBuilders, rego.Query(query))
	for _, path := range policyPaths {
		policyBytes, err := ioutil.ReadFile(path)
		if err != nil {
			return rego.PreparedEvalQuery{}, fmt.Errorf("policy file not found: %s", path)
		}
		regoBuilders = append(regoBuilders, rego.Module(path, string(policyBytes)))
	}
	for _, path := range dataPaths {
		dataBytes, err := ioutil.ReadFile(path)
		if err != nil {
			return rego.PreparedEvalQuery{}, fmt.Errorf("data file not found: %s", path)
		}
		regoBuilders = append(regoBuilders, rego.Module(path, string(dataBytes)))
	}

	evalQuery, err := rego.New(regoBuilders...).PrepareForEval(context.TODO())
	if err != nil {
		return rego.PreparedEvalQuery{}, fmt.Errorf("error creating evalQuery: %v", err)
	}
	return evalQuery, nil

}

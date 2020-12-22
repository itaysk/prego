package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/open-policy-agent/opa/rego"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Usage:  "Stream procesing using Rego",
		Action: appMain,
		Flags: []cli.Flag{
			&cli.StringSliceFlag{
				Name:     "files",
				Aliases:  []string{"f"},
				Required: true,
				Usage:    "Paths to load data and Rego modules from. Any file with a *.rego, *.yaml, or *.json extension will be loaded. The path can be a directory or a file, directories are loaded recursively.",
			},
			&cli.StringFlag{
				Name:     "package",
				Value:    "main",
				Required: false,
				Usage:    "Name of the Rego package to use (should match one the package in the provided file)",
			},
		},
	}
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func appMain(c *cli.Context) error {
	done := make(chan struct{})
	handleInterrupt(done)

	files := c.StringSlice("files")

	pkg := c.String("package")
	r := rego.New(
		rego.Load(files, nil),
		rego.Query(fmt.Sprintf("begin := data.%[1]s.BEGIN; end := data.%[1]s.END; main := data.%[1]s.MAIN", pkg)),
	)
	q, err := r.PrepareForEval(context.TODO())
	if err != nil {
		return err
	}

	begin, end, err := getFrame(q, pkg)
	if err != nil {
		return err
	}

	for _, l := range begin {
		fmt.Println(l)
	}

	scanner := bufio.NewScanner(os.Stdin)
	inputs := generate(*scanner, done)
	outputs := process(inputs, q, pkg)
	print(outputs)

	//block main
	<-done

	for _, l := range end {
		fmt.Println(l)
	}

	return nil
}

func handleInterrupt(done chan struct{}) {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		fmt.Println()
		close(done)
	}()
}

func getFrame(q rego.PreparedEvalQuery, pkg string) ([]interface{}, []interface{}, error) {
	results, err := q.Eval(context.TODO())
	if err != nil {
		return nil, nil, err
	}
	var begin, end []interface{}
	if len(results) > 0 {
		bs := results[0].Bindings
		if b, ok := bs["begin"]; ok {
			begin = b.([]interface{})
		}
		if e, ok := bs["end"]; ok {
			end = e.([]interface{})
		}
	}
	return begin, end, nil
}

func generate(in bufio.Scanner, done chan struct{}) <-chan map[string]interface{} {
	out := make(chan map[string]interface{})
	go func() {
		defer close(done)
		defer close(out)
		for in.Scan() {
			t := in.Text()
			var input map[string]interface{}
			json.Unmarshal([]byte(t), &input)
			out <- input
		}
	}()
	return out
}

func process(in <-chan map[string]interface{}, q rego.PreparedEvalQuery, pkg string) <-chan interface{} {
	out := make(chan interface{})
	go func() {
		defer close(out)
		for input := range in {
			results, err := q.Eval(context.TODO(), rego.EvalInput(input))
			if err != nil {
				continue
			}
			if len(results) > 0 {
				bs := results[0].Bindings
				if m, ok := bs["main"]; ok {
					for _, mi := range m.([]interface{}) {
						out <- mi
					}
				}
			}
		}
	}()
	return out
}

func print(in <-chan interface{}) {
	go func() {
		for i := range in {
			b, _ := json.Marshal(i)
			fmt.Println(string(b))
		}
	}()
}

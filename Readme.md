Prego - pipe into rego

Prego is a stream processing tool that uses Rego, the Open Policy Agent engine. It is meant to be used as the receiving end of a pipe, for example `cat myfile | prego`, similar to how you would use `sed`, `awk`, or `jq`.

## Getting Prego
Currently you have to build from source. Clone the repo and run `make build`.

## Basic Usage

Prego takes the following basic arguments:

- Policy files (rego files) with the `--policy` flag which takes a path to a file.
- Query to evaluate with the `--query` flag which takes a string (if ommitted it returns the entire `data` virtual document).
- Input to evaluate as a line from stdin.

Prego evaluates the loaded policy and picks the value of the last expression in each result in the result set. If there are multiple values to return, the return value is an array. This convention allows for multi part queries, while maintaining a simple API.

## Basic Example

```bash
cat test/testdata.jsonl | prego --policy test/example.rego
```

This will evaluate every line in `test/testdata.jsonl` using the policy `test/example.rego`, and print the evaluation result for each line into stdout.

```bash
cat test/testdata.jsonl | prego --policy test/example.rego --query 'data.example.myrule'
```

This will do the same as previously but each printed line will contain just the value of `myrule`.

```bash
cat test/testdata.jsonl | prego --policy test/example.rego --query 'data.example.hello' --print 'data.example.myrule'
```

This will print the value of `data.example.hello` from the loaded policy `test/example.rego` but only when the rule `data.example.myrule` evaluates to true

## Additional flags
You can specify the following additional flags:

- Data files (json files) with the `--data` flag which takes a path to a file. Currently data flag is the same as the policy flag.
- Stateful policies using the `--stateful` flag which is a boolean switch. See the Stateful Policies section below.
- Customize the output format using the `--output` flag which. See the Output section below.

## Output format

The following output formatters are supported:


- `json`: simple dump of the resulting object as JSON
- `regogo=query`: parse the result using [regogo](https://github.com/itaysk/regogo). e.g `--output 'regogo=input.expressions[0].value'`. input is every result in the Resultset.
-  `gotemplate=query`: parse the result using [go templates](https://golang.org/pkg/text/template/). e.g `--output 'gotemplate={{range .Expressions}}{{.Value}}{{end}}'`. input is every result in the Resultset. 

## Stateful Policies

You can build stateful policies using the `--stateful` flag. You use this flag in conjunction with a conventional policy file that have the followoing characteristics:
1. The package is `prego`
2. It defines a rule called `nextstate`
The result of `nextstate` is considered the "return value" of the stateful policy and is made available to the next evaluation under the variable `data.prego_state`.


### Stateful Example

```bash
cat test/testdata.jsonl | prego --query 'data.example2.return' --policy test/example2.rego --policy test/example2-sm.rego --stateful
```

This will evaulate every line in `test/testdata.jsonl` using the policy `test/example2.rego`, and the state machine defined in `test/example2-sm.rego`. The policy example2 is defining a rule which refers to the current state.

## How is this different from `opa eval`?

The opa binary has an `eval` subcommand that can evaluate policies from the command line. The primary difference, and the reason that prego exists, is that prego is made to process a stream of inputs, where 'opa eval' is made to process a single event at a time.  
You could wrapped 'opa eval' in some shell script that mimics the streaming behavior, but this would be bad for performance since for every event you pay for rebuilding the Rego context and for executing a new process. This might seem negligible but in a streaming scenario, where you pay this cost for every event, it sums up to noticable performance impact.

There are some additional differences:

- As an official, objective tool, 'opa eval' outputs the raw evaluation result. prego is trying to be more opinionated with outputting a meaningful result that can further piped to another tool.
- Related to the previous point, prego supports customizable output formatters.
- Prego supports stateful policies.

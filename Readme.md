Prego - pipe into rego

Prego is a stream processing tool that uses Rego, the Open Policy Agent engine. It is meant to be used as the receiving end of a pipe, for example `cat myfile | prego`, similar to how you would use `sed`, `awk`, or `jq`.

## Getting Prego
Currently you have to build from source. Clone the repo and run `make build`.

## Basic Usage

Prego takes the following basic arguments:

- Policy files (rego files) with the `--policy` flag which takes a path to a file
- Query to evaluate with the `--query` flag which takes a string (if ommitted it returns the entire `data` virtual document)
- Input to evaluate as a line from stdin

## Basic Example

```bash
cat test/testdata.jsonl | prego --policy test/example.rego
```

This will evaluate every line in `test/testdata.jsonl` using the policy `test/example.rego`, and print the evaluation result for each line into stdout.

```bash
cat test/testdata.jsonl | prego --policy test/example.rego --query 'data.example.myrule'
```

This will do the same as previously but each printed line will contain just the value of `myrule`.


## Additional flags
You can specify the following additional flags:

- Data files (json files) with the `--data` flag which takes a path to a file. Currently data flag is the same as the policy flag.
- Stateful policies using the `--stateful` flag which is a boolean switch. See the Stateful Policies section below.

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

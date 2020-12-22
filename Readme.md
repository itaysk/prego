Prego - Pipe into Rego

Prego is a stream processing cli tool that takes JSON from stdin, and processes it using [Rego](https://www.openpolicyagent.org/docs/latest/policy-language/) (the [Open Policy Agent](https://www.openpolicyagent.org/) language). It is inspired by [AWK](https://en.wikipedia.org/wiki/AWK) and can be thought of an AWK/Rego hybrid.

## Getting Prego
Currently you have to build from source. Clone the repo and run `make build`.

## Getting Started

Let's try Prego with the included example in [test/example.rego](test/example.rego). This is an example that:
1. Prints the time when it started and finished.
2. Looks in the input json for `.foo` elements which have a value of `bar`.
  1. Those items will be printed out.
3. Looks in the input json for `.hello` elements.
  1. For those items it will print the value of `.hello` in uppercase letters.
4. Every other items that doesn't match those conditions will be ignored (not printed).

Run `prego -f test/example.rego`, and type input documents (as JSON), or use the example data: `prego -f test/example.rego <test/testdata.jsonl`.

If we examine the Rego file, we see how it works:

```rego
package prego

BEGIN[out] {
  clock := time.clock(time.now_ns())
  out := sprintf("Started at %d:%d:%d", [clock[0], clock[1], clock[2]])
}

MAIN[out] {
  input.foo == "bar"
  out := input
}

MAIN[out] {
  out := upper(input.hello)
}

END[out] {
  clock := time.clock(time.now_ns())
  out := sprintf("Finished at %d:%d:%d", [clock[0], clock[1], clock[2]])
}
```

You "program" Prego by creating a Rego file with the "prego" package name, which includes the conventional rules: "BEGIN", "MAIN", and "END".
BEGIN is used to generate text before starting to process the input. END is used to generate text after the finishing to process the input. MAIN is used to process each item in the input.  
The rules are regular Rego rules, which evaluates to (i.e "returns") a set of strings. Because they are Rego rules they can leverage explicit and implicit assertions to conditionally execute the relevant rules. In the example, we have 2 implementations of the MAIN rule, but each one of them is handling a different case, and results in a different outcome.

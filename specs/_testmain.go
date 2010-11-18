package main

import "console/specs"
import "testing"
import __regexp__ "regexp"

var tests = []testing.InternalTest{
	{"specs.TestConsoleSpecs", specs.TestConsoleSpecs},
}
var benchmarks = []testing.InternalBenchmark{}

func main() {
	testing.Main(__regexp__.MatchString, tests)
	testing.RunBenchmarks(__regexp__.MatchString, benchmarks)
}

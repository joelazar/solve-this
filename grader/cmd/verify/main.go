package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"

	"github.com/joelazar/solve-this-grader/internal/bugs"
	"github.com/joelazar/solve-this-grader/internal/run"
	"github.com/joelazar/solve-this-grader/internal/score"
)

func main() {
	home, _ := os.UserHomeDir()
	grader := flag.String("grader", ".", "grader repository holding the hidden tests")
	src := flag.String("src", filepath.Join(home, "Code/joelazar/solve-this"), "baseline repository")
	rev := flag.String("rev", "HEAD", "baseline revision")
	flag.Parse()

	work, err := os.MkdirTemp("", "solve-this-verify")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(work)

	revision, err := run.Git(*src, "rev-parse", *rev)
	if err != nil {
		log.Fatal(err)
	}

	failures := 0
	failures += check(*grader, *src, revision, work, "baseline", nil)
	for _, bug := range bugs.All {
		failures += check(*grader, *src, revision, work, bug.ID, []bugs.Bug{bug})
	}
	if failures > 0 {
		fmt.Printf("\n%d variants behaved unexpectedly\n", failures)
		os.Exit(1)
	}
	fmt.Printf("\n%d bugs verified, each isolated to its own test\n", len(bugs.All))
}

func check(grader, src, revision, work, name string, selected []bugs.Bug) int {
	dir := filepath.Join(work, name)
	if err := run.Export(src, revision, dir); err != nil {
		log.Fatal(err)
	}
	for _, bug := range selected {
		if err := bugs.Apply(dir, bug); err != nil {
			log.Fatal(err)
		}
	}
	if files := bugs.Files(selected); len(files) > 0 {
		if _, err := run.Command(dir, nil, "gofmt", append([]string{"-w"}, files...)...); err != nil {
			log.Fatal(err)
		}
	}
	if err := run.Init(dir); err != nil {
		log.Fatal(err)
	}

	report, err := score.Evaluate(grader, dir)
	if err != nil {
		log.Fatal(err)
	}

	expected := map[string]bool{}
	for _, bug := range selected {
		expected[bug.Test] = true
	}
	var unexpected, missing []string
	for test, passed := range report.Tests {
		if !passed && !expected[test] {
			unexpected = append(unexpected, test)
		}
		if passed && expected[test] {
			missing = append(missing, test)
		}
	}
	sort.Strings(unexpected)
	sort.Strings(missing)

	status := "ok"
	problems := 0
	if !report.Build || !report.Vet {
		status = fmt.Sprintf("build %v vet %v", report.Build, report.Vet)
		problems++
	}
	if len(unexpected) > 0 {
		status = "also failing: " + fmt.Sprint(unexpected)
		problems++
	}
	if len(missing) > 0 {
		status = "not detected: " + fmt.Sprint(missing)
		problems++
	}
	fmt.Printf("%-22s %d tests  %s\n", name, len(report.Tests), status)
	return problems
}

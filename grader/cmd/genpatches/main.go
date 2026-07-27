package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/joelazar/solve-this-grader/internal/bugs"
	"github.com/joelazar/solve-this-grader/internal/run"
)

func main() {
	home, _ := os.UserHomeDir()
	src := flag.String("src", filepath.Join(home, "Code/joelazar/solve-this"), "baseline repository")
	rev := flag.String("rev", "HEAD", "baseline revision")
	out := flag.String("out", "bugs", "directory for generated patches")
	flag.Parse()

	work, err := os.MkdirTemp("", "solve-this-patches")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(work)

	if err := run.Export(*src, *rev, work); err != nil {
		log.Fatal(err)
	}
	if err := run.Init(work); err != nil {
		log.Fatal(err)
	}
	if err := os.MkdirAll(*out, 0o755); err != nil {
		log.Fatal(err)
	}

	for _, bug := range bugs.All {
		if err := bugs.Apply(work, bug); err != nil {
			log.Fatal(err)
		}
		if _, err := run.Command(work, nil, "gofmt", append([]string{"-w"}, bugs.Files([]bugs.Bug{bug})...)...); err != nil {
			log.Fatal(err)
		}
		diff, err := run.Git(work, "diff")
		if err != nil {
			log.Fatal(err)
		}
		header := fmt.Sprintf("tier %d, %s\nbreaks: %s\ncovered by: %s\n\n", bug.Tier, bug.Class, bug.Spec, bug.Test)
		path := filepath.Join(*out, bug.ID+".patch")
		if err := os.WriteFile(path, []byte(header+diff+"\n"), 0o644); err != nil {
			log.Fatal(err)
		}
		if _, err := run.Git(work, "checkout", "--", "."); err != nil {
			log.Fatal(err)
		}
		fmt.Println(path)
	}
}

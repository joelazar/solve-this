package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/joelazar/solve-this/grader/internal/run"
	"github.com/joelazar/solve-this/grader/internal/score"
)

func main() {
	root, err := run.Root()
	if err != nil {
		log.Fatal(err)
	}
	grader := flag.String("grader", filepath.Join(root, "grader"), "directory holding the hidden tests")
	manifests := flag.String("manifests", filepath.Join(root, "runs"), "directory holding run manifests")
	id := flag.String("id", "", "run id")
	flag.Parse()

	if *id == "" {
		log.Fatal("-id is required")
	}
	body, err := os.ReadFile(filepath.Join(*manifests, *id+".json"))
	if err != nil {
		log.Fatal(err)
	}
	var manifest run.Manifest
	if err := json.Unmarshal(body, &manifest); err != nil {
		log.Fatal(err)
	}

	report, err := score.Evaluate(*grader, manifest.Dir)
	if err != nil {
		log.Fatal(err)
	}
	card := score.Build(manifest.ID, manifest.Dir, manifest.Bugs, report)

	encoded, _ := json.MarshalIndent(card, "", "  ")
	path := filepath.Join(*manifests, *id+"-score.json")
	if err := os.WriteFile(path, append(encoded, '\n'), 0o644); err != nil {
		log.Fatal(err)
	}
	fmt.Println(card)
}

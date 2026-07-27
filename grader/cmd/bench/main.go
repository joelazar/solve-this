package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/joelazar/solve-this/grader/internal/bugs"
	"github.com/joelazar/solve-this/grader/internal/run"
	"github.com/joelazar/solve-this/grader/internal/score"
	"github.com/joelazar/solve-this/grader/internal/variant"
)

func main() {
	root, err := run.Root()
	if err != nil {
		log.Fatal(err)
	}
	src := flag.String("src", root, "repository holding the app/ baseline")
	rev := flag.String("rev", "HEAD", "baseline revision")
	out := flag.String("out", filepath.Join(filepath.Dir(root), "solve-this-runs"), "directory holding run trees, kept outside the repository")
	manifests := flag.String("manifests", filepath.Join(root, "runs"), "directory holding run artifacts")
	grader := flag.String("grader", filepath.Join(root, "grader"), "directory holding the hidden tests")
	id := flag.String("id", "", "run id, defaults to a timestamp")
	selection := flag.String("bugs", "", "comma separated bug ids, empty means random")
	n := flag.Int("n", len(bugs.All), "number of bugs when selecting randomly")
	seed := flag.Int64("seed", time.Now().UnixNano(), "selection seed")
	tiersFlag := flag.String("tiers", "1,2", "comma separated tiers for random selection, ignored with -bugs")
	mode := flag.String("mode", "hunt", "prompt template in prompts/: hunt, count or report")
	promptPath := flag.String("prompt", "", "custom prompt template, overrides -mode")
	agent := flag.String("agent", "", "agent command, run through sh -c inside the run directory; empty prints the prompt for a manual session")
	flag.Parse()

	if *id == "" {
		*id = time.Now().Format("20060102-150405")
	}

	var ids []string
	if *selection != "" {
		ids = strings.Split(*selection, ",")
	}
	var tiers []int
	for _, part := range strings.Split(*tiersFlag, ",") {
		tier, err := strconv.Atoi(strings.TrimSpace(part))
		if err != nil {
			log.Fatalf("invalid -tiers value %q", *tiersFlag)
		}
		tiers = append(tiers, tier)
	}
	manifest, err := variant.Create(variant.Options{
		Src: *src, Rev: *rev, Out: *out, Manifests: *manifests,
		ID: *id, Seed: *seed, N: *n, Tiers: tiers, IDs: ids,
	})
	if err != nil {
		log.Fatal(err)
	}

	tmpl := *promptPath
	if tmpl == "" {
		tmpl = filepath.Join(*grader, "prompts", *mode+".md")
	}
	prompt, err := render(tmpl, manifest)
	if err != nil {
		log.Fatal(err)
	}
	promptFile, err := filepath.Abs(filepath.Join(*manifests, *id+"-prompt.md"))
	if err != nil {
		log.Fatal(err)
	}
	if err := os.WriteFile(promptFile, []byte(prompt), 0o644); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("run dir  %s\nbugs     %d: %s\nprompt   %s\n", manifest.Dir, len(manifest.Bugs), strings.Join(manifest.Bugs, " "), promptFile)

	if *agent == "" {
		fmt.Printf("\n%s\nrun your agent in the run dir, then: mise run score -- -id %s\n", prompt, *id)
		return
	}

	logFile, err := os.Create(filepath.Join(*manifests, *id+"-agent.log"))
	if err != nil {
		log.Fatal(err)
	}
	sink := io.MultiWriter(os.Stdout, logFile)
	cmd := exec.Command("sh", "-c", *agent)
	cmd.Dir = manifest.Dir
	cmd.Env = append(os.Environ(), "PROMPT="+prompt, "PROMPT_FILE="+promptFile)
	cmd.Stdout = sink
	cmd.Stderr = sink

	started := time.Now()
	agentErr := cmd.Run()
	elapsed := time.Since(started)
	logFile.Close()
	if agentErr != nil {
		log.Printf("agent exited with an error: %v", agentErr)
	}

	report, err := score.Evaluate(*grader, manifest.Dir)
	if err != nil {
		log.Fatal(err)
	}
	card := score.Build(manifest.ID, manifest.Dir, manifest.Bugs, report)
	card.Agent = *agent
	card.Seconds = elapsed.Seconds()

	encoded, _ := json.MarshalIndent(card, "", "  ")
	path := filepath.Join(*manifests, *id+"-score.json")
	if err := os.WriteFile(path, append(encoded, '\n'), 0o644); err != nil {
		log.Fatal(err)
	}
	fmt.Println(card)
}

func render(path string, manifest run.Manifest) (string, error) {
	tmpl, err := template.ParseFiles(path)
	if err != nil {
		return "", err
	}
	data := struct {
		N        int
		Symptoms []string
	}{N: len(manifest.Bugs)}
	for _, id := range manifest.Bugs {
		if bug, ok := bugs.ByID(id); ok {
			data.Symptoms = append(data.Symptoms, bug.Symptom)
		}
	}
	var rendered bytes.Buffer
	if err := tmpl.Execute(&rendered, data); err != nil {
		return "", err
	}
	return rendered.String(), nil
}

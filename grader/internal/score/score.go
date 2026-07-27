package score

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"slices"
	"strconv"
	"strings"

	"github.com/joelazar/solve-this/grader/internal/bugs"
	"github.com/joelazar/solve-this/grader/internal/run"
)

type Report struct {
	Build bool            `json:"build"`
	Vet   bool            `json:"vet"`
	Tests map[string]bool `json:"tests"`
}

type Diff struct {
	Files   int `json:"files"`
	Added   int `json:"added"`
	Removed int `json:"removed"`
}

type Outcome struct {
	ID    string `json:"id"`
	Tier  int    `json:"tier"`
	Class string `json:"class"`
	Fixed bool   `json:"fixed"`
}

type Card struct {
	Run          string    `json:"run"`
	Dir          string    `json:"dir"`
	Agent        string    `json:"agent,omitempty"`
	Seconds      float64   `json:"seconds,omitempty"`
	Build        bool      `json:"build"`
	Vet          bool      `json:"vet"`
	Planted      int       `json:"planted"`
	Fixed        int       `json:"fixed"`
	Bugs         []Outcome `json:"bugs"`
	DecoysBroken []string  `json:"decoys_broken"`
	Collateral   []string  `json:"collateral"`
	Regressions  []string  `json:"regressions"`
	Diff         Diff      `json:"diff"`
}

func Evaluate(grader, dir string) (Report, error) {
	report := Report{Tests: map[string]bool{}}
	if _, err := run.Command(dir, nil, "go", "build", "./..."); err != nil {
		return report, nil
	}
	report.Build = true
	if _, err := run.Command(dir, nil, "go", "vet", "./..."); err == nil {
		report.Vet = true
	}

	cmd := exec.Command("go", "test", "-json", "-count=1", "./tests/...")
	cmd.Dir = grader
	cmd.Env = append(os.Environ(), "SOLVE_THIS_DIR="+dir)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	cmd.Run()

	scanner := bufio.NewScanner(bytes.NewReader(out.Bytes()))
	scanner.Buffer(make([]byte, 0, 1<<20), 1<<20)
	for scanner.Scan() {
		var event struct {
			Action string
			Test   string
		}
		if err := json.Unmarshal(scanner.Bytes(), &event); err != nil {
			continue
		}
		if event.Test == "" || strings.Contains(event.Test, "/") {
			continue
		}
		switch event.Action {
		case "pass":
			report.Tests[event.Test] = true
		case "fail":
			report.Tests[event.Test] = false
		}
	}
	return report, scanner.Err()
}

func Build(id, dir string, planted []string, report Report) Card {
	applied := map[string]bool{}
	for _, bug := range planted {
		applied[bug] = true
	}
	card := Card{Run: id, Dir: dir, Build: report.Build, Vet: report.Vet, Planted: len(planted)}
	for _, bug := range bugs.All {
		if !applied[bug.ID] {
			if passed, ok := report.Tests[bug.Test]; ok && !passed {
				card.Collateral = append(card.Collateral, bug.ID)
			}
			continue
		}
		fixed := report.Tests[bug.Test]
		card.Bugs = append(card.Bugs, Outcome{ID: bug.ID, Tier: bug.Tier, Class: bug.Class, Fixed: fixed})
		if fixed {
			card.Fixed++
		}
	}
	for _, decoy := range bugs.Decoys {
		if passed, ok := report.Tests[decoy.Test]; ok && !passed {
			card.DecoysBroken = append(card.DecoysBroken, decoy.ID)
		}
	}
	for name, passed := range report.Tests {
		if !passed && strings.HasPrefix(name, "TestRegression") {
			card.Regressions = append(card.Regressions, name)
		}
	}
	slices.Sort(card.Regressions)
	card.Diff = diff(dir)
	return card
}

func (c Card) String() string {
	var b strings.Builder
	fmt.Fprintf(&b, "run       %s\n", c.Run)
	if c.Agent != "" {
		fmt.Fprintf(&b, "agent     %s\n", c.Agent)
	}
	if c.Seconds > 0 {
		fmt.Fprintf(&b, "time      %.0fs\n", c.Seconds)
	}
	fmt.Fprintf(&b, "build     %s\n", mark(c.Build))
	fmt.Fprintf(&b, "vet       %s\n", mark(c.Vet))
	fmt.Fprintf(&b, "fixed     %d/%d\n", c.Fixed, c.Planted)
	for _, bug := range c.Bugs {
		fmt.Fprintf(&b, "  %s tier %d  %-22s %s\n", mark(bug.Fixed), bug.Tier, bug.ID, bug.Class)
	}
	fmt.Fprintf(&b, "decoys    %s\n", list(c.DecoysBroken, "intact"))
	fmt.Fprintf(&b, "collateral %s\n", list(c.Collateral, "none"))
	fmt.Fprintf(&b, "regressions %s\n", list(c.Regressions, "none"))
	fmt.Fprintf(&b, "diff      %d files +%d -%d", c.Diff.Files, c.Diff.Added, c.Diff.Removed)
	return b.String()
}

func mark(ok bool) string {
	if ok {
		return "+"
	}
	return "-"
}

func list(values []string, empty string) string {
	if len(values) == 0 {
		return empty
	}
	return strings.Join(values, " ")
}

func diff(dir string) Diff {
	root, err := run.Git(dir, "rev-list", "--max-parents=0", "HEAD")
	if err != nil {
		return Diff{}
	}
	numstat, err := run.Git(dir, "diff", "--numstat", root)
	if err != nil {
		return Diff{}
	}
	var stat Diff
	for _, line := range strings.Split(numstat, "\n") {
		fields := strings.Fields(line)
		if len(fields) != 3 {
			continue
		}
		stat.Files++
		added, _ := strconv.Atoi(fields[0])
		removed, _ := strconv.Atoi(fields[1])
		stat.Added += added
		stat.Removed += removed
	}
	return stat
}

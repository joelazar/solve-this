package variant

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/joelazar/solve-this/grader/internal/bugs"
	"github.com/joelazar/solve-this/grader/internal/run"
)

type Options struct {
	Src       string
	Rev       string
	Out       string
	Manifests string
	ID        string
	Seed      int64
	N         int
	IDs       []string
}

func Create(o Options) (run.Manifest, error) {
	dir := filepath.Join(o.Out, o.ID)
	if _, err := os.Stat(dir); err == nil {
		return run.Manifest{}, fmt.Errorf("%s already exists", dir)
	}

	selected := bugs.Select(o.Seed, o.N)
	if len(o.IDs) > 0 {
		selected = nil
		for _, want := range o.IDs {
			bug, ok := bugs.ByID(strings.TrimSpace(want))
			if !ok {
				return run.Manifest{}, fmt.Errorf("unknown bug %q", want)
			}
			selected = append(selected, bug)
		}
	}

	revision, err := run.Git(o.Src, "rev-parse", o.Rev)
	if err != nil {
		return run.Manifest{}, err
	}
	if err := run.Export(o.Src, revision, dir); err != nil {
		return run.Manifest{}, err
	}

	ids := make([]string, 0, len(selected))
	for _, bug := range selected {
		if err := bugs.Apply(dir, bug); err != nil {
			return run.Manifest{}, err
		}
		ids = append(ids, bug.ID)
	}
	if files := bugs.Files(selected); len(files) > 0 {
		if _, err := run.Command(dir, nil, "gofmt", append([]string{"-w"}, files...)...); err != nil {
			return run.Manifest{}, err
		}
	}
	if output, err := run.Command(dir, nil, "go", "build", "./..."); err != nil {
		return run.Manifest{}, fmt.Errorf("variant does not build: %s", output)
	}
	if err := run.Init(dir); err != nil {
		return run.Manifest{}, err
	}

	manifest := run.Manifest{
		ID:       o.ID,
		Created:  time.Now().Format(time.RFC3339),
		Source:   o.Src,
		Revision: revision,
		Seed:     o.Seed,
		Dir:      dir,
		Bugs:     ids,
	}
	if err := os.MkdirAll(o.Manifests, 0o755); err != nil {
		return run.Manifest{}, err
	}
	body, _ := json.MarshalIndent(manifest, "", "  ")
	if err := os.WriteFile(filepath.Join(o.Manifests, o.ID+".json"), append(body, '\n'), 0o644); err != nil {
		return run.Manifest{}, err
	}
	return manifest, nil
}

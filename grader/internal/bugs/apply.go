package bugs

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

func Apply(dir string, bug Bug) error {
	for _, edit := range bug.Edits {
		path := filepath.Join(dir, edit.File)
		src, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("%s: %w", bug.ID, err)
		}
		if count := strings.Count(string(src), edit.Old); count != 1 {
			return fmt.Errorf("%s: anchor matches %d times in %s", bug.ID, count, edit.File)
		}
		patched := strings.Replace(string(src), edit.Old, edit.New, 1)
		if err := os.WriteFile(path, []byte(patched), 0o644); err != nil {
			return fmt.Errorf("%s: %w", bug.ID, err)
		}
	}
	return nil
}

func Files(selected []Bug) []string {
	seen := map[string]struct{}{}
	var files []string
	for _, bug := range selected {
		for _, edit := range bug.Edits {
			if _, ok := seen[edit.File]; ok {
				continue
			}
			seen[edit.File] = struct{}{}
			files = append(files, edit.File)
		}
	}
	return files
}

func Select(seed int64, n int, tiers []int) []Bug {
	shuffled := make([]Bug, 0, len(All))
	for _, bug := range All {
		if len(tiers) == 0 || slices.Contains(tiers, bug.Tier) {
			shuffled = append(shuffled, bug)
		}
	}
	rng := rand.New(rand.NewSource(seed))
	rng.Shuffle(len(shuffled), func(i, j int) {
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	})
	if n > len(shuffled) {
		n = len(shuffled)
	}
	picked := shuffled[:n]
	ordered := make([]Bug, 0, len(picked))
	for _, bug := range All {
		for _, p := range picked {
			if p.ID == bug.ID {
				ordered = append(ordered, bug)
				break
			}
		}
	}
	return ordered
}

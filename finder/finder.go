// Package finder searches the filesystem for the launcher's file mode. It
// shells out to fd when available and falls back to a bounded directory walk.
package finder

import (
	"bufio"
	"context"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	DefaultLimit = 40
	walkBudget   = 200_000 // max entries visited by the fallback walker
	searchLimit  = 3 * time.Second
)

// Roots are the directories searched for non-path queries.
var Roots = func() []string {
	home, err := os.UserHomeDir()
	if err != nil {
		return []string{"/"}
	}
	return []string{home}
}()

// Excludes are directory names skipped during search.
var Excludes = []string{
	".git", "node_modules", ".cache", ".direnv", "target", ".npm", ".cargo",
	".rustup", ".local/share/Trash", "__pycache__", ".venv", "venv", ".gradle",
	".m2", "go/pkg", ".steam", ".var",
}

// Result is a single filesystem match.
type Result struct {
	Path  string
	IsDir bool
}

// Name returns the display name of the entry.
func (r Result) Name() string {
	name := filepath.Base(r.Path)
	if r.IsDir {
		return name + "/"
	}
	return name
}

// Dir returns the parent directory with $HOME abbreviated to ~.
func (r Result) Dir() string {
	return Abbreviate(filepath.Dir(r.Path))
}

// Abbreviate replaces the home directory prefix with ~.
func Abbreviate(path string) string {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return path
	}
	if path == home {
		return "~"
	}
	if strings.HasPrefix(path, home+"/") {
		return "~" + path[len(home):]
	}
	return path
}

// Expand turns ~ prefixes into the home directory.
func Expand(path string) string {
	if path == "~" || strings.HasPrefix(path, "~/") {
		if home, err := os.UserHomeDir(); err == nil {
			return home + path[1:]
		}
	}
	return path
}

// Search finds files and directories matching query. A query containing a
// slash is treated as a path and completed against its parent directory;
// anything else is a case-insensitive name search under Roots.
func Search(query string, limit int) []Result {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil
	}
	if limit <= 0 {
		limit = DefaultLimit
	}
	if strings.ContainsRune(query, '/') || strings.HasPrefix(query, "~") {
		return completePath(Expand(query), limit)
	}
	if _, err := exec.LookPath("fd"); err == nil {
		if res, ok := searchFd(query, Roots, limit); ok {
			return res
		}
	}
	return searchWalk(query, Roots, limit)
}

// completePath lists entries in the directory part of path whose names begin
// with the remaining partial name (like shell tab-completion).
func completePath(path string, limit int) []Result {
	dir, partial := filepath.Split(path)
	if dir == "" {
		dir = "."
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	lowPartial := strings.ToLower(partial)
	var results []Result
	for _, e := range entries {
		name := e.Name()
		if partial == "" && strings.HasPrefix(name, ".") {
			continue
		}
		if !strings.HasPrefix(strings.ToLower(name), lowPartial) {
			continue
		}
		results = append(results, Result{Path: filepath.Join(dir, name), IsDir: e.IsDir()})
	}
	sort.Slice(results, func(i, j int) bool {
		if results[i].IsDir != results[j].IsDir {
			return results[i].IsDir
		}
		return results[i].Path < results[j].Path
	})
	if len(results) > limit {
		results = results[:limit]
	}
	return results
}

func searchFd(query string, roots []string, limit int) ([]Result, bool) {
	ctx, cancel := context.WithTimeout(context.Background(), searchLimit)
	defer cancel()

	args := []string{
		"--fixed-strings", "--ignore-case", "--hidden", "--absolute-path",
		"--color", "never", "--max-results", strconv.Itoa(limit * 4),
	}
	for _, ex := range Excludes {
		args = append(args, "--exclude", ex)
	}
	args = append(args, query)
	args = append(args, roots...)

	cmd := exec.CommandContext(ctx, "fd", args...)
	out, err := cmd.StdoutPipe()
	if err != nil {
		return nil, false
	}
	if err := cmd.Start(); err != nil {
		return nil, false
	}

	var results []Result
	scanner := bufio.NewScanner(out)
	for scanner.Scan() {
		p := strings.TrimRight(scanner.Text(), "/")
		if p == "" {
			continue
		}
		info, err := os.Stat(p)
		isDir := err == nil && info.IsDir()
		results = append(results, Result{Path: p, IsDir: isDir})
	}
	_ = cmd.Wait()
	if ctx.Err() != nil && len(results) == 0 {
		return nil, false
	}
	return rank(results, query, limit), true
}

func searchWalk(query string, roots []string, limit int) []Result {
	lowQuery := strings.ToLower(query)
	deadline := time.Now().Add(searchLimit)
	visited := 0
	var results []Result

	for _, root := range roots {
		_ = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return nil
			}
			visited++
			if visited > walkBudget || time.Now().After(deadline) || len(results) >= limit*4 {
				return filepath.SkipAll
			}
			name := d.Name()
			if d.IsDir() && path != root {
				for _, ex := range Excludes {
					if name == ex || strings.HasSuffix(path, "/"+ex) {
						return filepath.SkipDir
					}
				}
			}
			if strings.Contains(strings.ToLower(name), lowQuery) {
				results = append(results, Result{Path: path, IsDir: d.IsDir()})
			}
			return nil
		})
	}
	return rank(results, query, limit)
}

// rank orders matches so that exact and prefix name matches come first, then
// shorter paths (which are usually more relevant than deeply nested ones).
func rank(results []Result, query string, limit int) []Result {
	lowQuery := strings.ToLower(query)
	score := func(r Result) int {
		name := strings.ToLower(filepath.Base(r.Path))
		s := strings.Count(r.Path, "/")
		switch {
		case name == lowQuery:
			s -= 1000
		case strings.HasPrefix(name, lowQuery):
			s -= 500
		case strings.Contains(name, lowQuery):
			s -= 100
		}
		if strings.HasPrefix(filepath.Base(r.Path), ".") {
			s += 50
		}
		return s
	}
	sort.SliceStable(results, func(i, j int) bool {
		si, sj := score(results[i]), score(results[j])
		if si != sj {
			return si < sj
		}
		return results[i].Path < results[j].Path
	})
	if len(results) > limit {
		results = results[:limit]
	}
	return results
}

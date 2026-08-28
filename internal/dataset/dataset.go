package dataset

import (
	"bufio"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type CategoryFile struct {
	Name string
	Path string
}

func List(dir string) ([]CategoryFile, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var out []CategoryFile
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		out = append(out, CategoryFile{
			Name: e.Name(),
			Path: filepath.Join(dir, e.Name()),
		})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, nil
}

func ReadLines(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var lines []string
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 1024*1024), 10*1024*1024)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		lines = append(lines, line)
	}
	return lines, sc.Err()
}

func HasEntries(path string) (bool, error) {
	lines, err := ReadLines(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return len(lines) > 0, nil
}

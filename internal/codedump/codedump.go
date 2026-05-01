package codedump

import (
	"bytes"
	"fmt"
	"io"
	"os"
	pathpkg "path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/go-git/go-billy/v5/osfs"
	"github.com/go-git/go-git/v5/plumbing/format/gitignore"
)

type dirNode struct {
	name    string
	subdirs []*dirNode
	files   []string
}

func Run(root string) (treeText string, relFiles []string, err error) {
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return "", nil, err
	}
	fi, err := os.Stat(absRoot)
	if err != nil {
		return "", nil, err
	}
	if !fi.IsDir() {
		return "", nil, fmt.Errorf("input is not a directory: %s", absRoot)
	}

	patterns, err := gitignore.ReadPatterns(osfs.New(absRoot), nil)
	if err != nil {
		return "", nil, fmt.Errorf("read gitignore patterns: %w", err)
	}
	matcher := gitignore.NewMatcher(patterns)

	rootNode := &dirNode{name: filepath.Base(absRoot)}
	relFiles, err = buildTree(absRoot, rootNode, "", matcher)
	if err != nil {
		return "", nil, err
	}
	return formatTree(rootNode), relFiles, nil
}

func buildTree(absDir string, n *dirNode, rel string, matcher gitignore.Matcher) ([]string, error) {
	names, err := readSortedNames(absDir)
	if err != nil {
		return nil, err
	}
	var out []string
	for _, name := range names {
		if name == ".git" {
			continue
		}
		childAbs := filepath.Join(absDir, name)
		fi, err := os.Lstat(childAbs)
		if err != nil {
			return nil, err
		}
		if fi.Mode()&os.ModeSymlink != 0 {
			continue
		}
		relSeg := name
		if rel != "" {
			relSeg = rel + "/" + name
		}
		segs := strings.Split(relSeg, "/")
		isDir := fi.IsDir()
		if matcher.Match(segs, isDir) {
			continue
		}
		if isDir {
			sub := &dirNode{name: name}
			n.subdirs = append(n.subdirs, sub)
			subFiles, err := buildTree(childAbs, sub, relSeg, matcher)
			if err != nil {
				return nil, err
			}
			out = append(out, subFiles...)
		} else {
			n.files = append(n.files, name)
			out = append(out, filepath.ToSlash(relSeg))
		}
	}
	sort.Strings(n.files)
	sort.Slice(n.subdirs, func(i, j int) bool { return n.subdirs[i].name < n.subdirs[j].name })
	return out, nil
}

func readSortedNames(dir string) ([]string, error) {
	f, err := os.Open(dir)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	names, err := f.Readdirnames(-1)
	if err != nil {
		return nil, err
	}
	sort.Strings(names)
	return names, nil
}

func formatTree(root *dirNode) string {
	var b strings.Builder
	b.WriteString(root.name)
	b.WriteByte('/')
	b.WriteByte('\n')
	items := treeItems(root)
	writeTreeLines(&b, items, "")
	return b.String()
}

type treeItem struct {
	name  string
	isDir bool
	node  *dirNode
}

func treeItems(n *dirNode) []treeItem {
	var items []treeItem
	for _, d := range n.subdirs {
		items = append(items, treeItem{name: d.name, isDir: true, node: d})
	}
	for _, f := range n.files {
		items = append(items, treeItem{name: f, isDir: false, node: nil})
	}
	sort.Slice(items, func(i, j int) bool { return items[i].name < items[j].name })
	return items
}

func writeTreeLines(b *strings.Builder, items []treeItem, prefix string) {
	for i, it := range items {
		last := i == len(items)-1
		branch := "├── "
		nextPrefix := prefix + "│   "
		if last {
			branch = "└── "
			nextPrefix = prefix + "    "
		}
		b.WriteString(prefix)
		b.WriteString(branch)
		b.WriteString(it.name)
		if it.isDir {
			b.WriteByte('/')
		}
		b.WriteByte('\n')
		if it.isDir && it.node != nil {
			writeTreeLines(b, treeItems(it.node), nextPrefix)
		}
	}
}

func WriteDump(w io.Writer, root string, relFiles []string) error {
	for _, rel := range relFiles {
		rel = filepath.ToSlash(rel)
		abs := filepath.Join(root, filepath.FromSlash(rel))
		data, err := os.ReadFile(abs)
		if err != nil {
			return fmt.Errorf("read %s: %w", rel, err)
		}
		if bytes.IndexByte(data, 0) >= 0 {
			continue
		}
		if _, err := fmt.Fprintf(w, "%s\n", rel); err != nil {
			return err
		}
		if _, err := w.Write(data); err != nil {
			return err
		}
		if len(data) > 0 && data[len(data)-1] != '\n' {
			if _, err := w.Write([]byte{'\n'}); err != nil {
				return err
			}
		}
	}
	return nil
}

func SafeOutputName(name string) (string, error) {
	if name == "" {
		return "", fmt.Errorf("name must not be empty")
	}
	if strings.ContainsAny(name, `/\:`) {
		return "", fmt.Errorf("invalid name")
	}
	base := pathpkg.Base(name)
	if base != name || name == "." || name == ".." {
		return "", fmt.Errorf("invalid name")
	}
	return name, nil
}

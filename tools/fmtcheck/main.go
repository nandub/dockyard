package main

import (
	"bytes"
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	roots := os.Args[1:]
	if len(roots) == 0 {
		roots = []string{"./cmd", "./internal", "./tools"}
	}

	var unformatted []string

	for _, root := range roots {
		if err := walk(root, &unformatted); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(2)
		}
	}

	if len(unformatted) > 0 {
		for _, path := range unformatted {
			fmt.Println(path)
		}
		os.Exit(1)
	}
}

func walk(root string, unformatted *[]string) error {
	info, err := os.Stat(root)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("stat %s: %w", root, err)
	}

	if !info.IsDir() {
		return checkFile(root, unformatted)
	}

	return filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if entry.IsDir() {
			name := entry.Name()
			if strings.HasPrefix(name, ".") || name == "vendor" || name == "bin" {
				return filepath.SkipDir
			}
			return nil
		}

		if filepath.Ext(path) != ".go" {
			return nil
		}

		return checkFile(path, unformatted)
	})
}

func checkFile(path string, unformatted *[]string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read %s: %w", path, err)
	}

	formatted, err := format.Source(data)
	if err != nil {
		return fmt.Errorf("format %s: %w", path, err)
	}

	if !bytes.Equal(data, formatted) {
		*unformatted = append(*unformatted, filepath.Clean(path))
	}

	return nil
}

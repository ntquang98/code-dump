package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"code-dump/internal/codedump"

	"github.com/spf13/pflag"
)

func main() {
	pflag.Usage = func() {
		fmt.Fprintln(os.Stderr, "usage: codedump --input <path-to-project> --name <name>")
		pflag.PrintDefaults()
	}
	var input, name string
	pflag.StringVar(&input, "input", "", "path to project root directory")
	pflag.StringVar(&name, "name", "", "output basename token")
	pflag.Parse()

	if input == "" || name == "" {
		pflag.Usage()
		os.Exit(2)
	}
	name, err := codedump.SafeOutputName(name)
	if err != nil {
		log.Fatalf("name: %v", err)
	}

	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	treePath := filepath.Join(cwd, fmt.Sprintf("src_tree_%s.txt", name))
	dumpPath := filepath.Join(cwd, fmt.Sprintf("%s_dump.txt", name))

	treeText, relFiles, err := codedump.Run(input)
	if err != nil {
		log.Fatal(err)
	}

	if err := os.WriteFile(treePath, []byte(treeText), 0o644); err != nil {
		log.Fatalf("write tree: %v", err)
	}

	dumpFile, err := os.Create(dumpPath)
	if err != nil {
		log.Fatalf("create dump: %v", err)
	}
	defer dumpFile.Close()

	absInput, err := filepath.Abs(input)
	if err != nil {
		log.Fatal(err)
	}
	if err := codedump.WriteDump(dumpFile, absInput, relFiles); err != nil {
		log.Fatalf("write dump: %v", err)
	}

	fmt.Println(treePath)
	fmt.Println(dumpPath)
}

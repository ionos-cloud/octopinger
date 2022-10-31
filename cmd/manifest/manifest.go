package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/pflag"
)

type flags struct {
	Files  []string
	Output string
}

func main() {
	log.SetFlags(0)
	log.SetOutput(os.Stderr)

	f := &flags{}

	pflag.StringSliceVar(&f.Files, "file", f.Files, "files")
	pflag.StringVar(&f.Output, "output", f.Output, "output")
	pflag.Parse()

	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	ss := make([]string, 0)

	for _, f := range f.Files {
		m, err := os.ReadFile(filepath.Clean(filepath.Join(cwd, f)))
		if err != nil {
			log.Fatal(err)
		}

		ss = append(ss, fmt.Sprintf("---\n%s", string(m)))
	}

	bb := []byte(strings.Join(ss, ""))
	err = os.WriteFile(filepath.Clean(filepath.Join(cwd, f.Output)), bb, 0600)
	if err != nil {
		log.Fatal(err)
	}
}

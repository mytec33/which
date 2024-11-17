package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

const (
	EXIT_SUCCESS       = 0
	EXIT_INVALID_ARGS  = 1
	EXIT_NOT_ALL_FOUND = 2
	EXIT_PATH_EMPTY    = 3
)

func main() {
	flag.Usage = printUsage

	var aFlag = flag.Bool("a", false, "list all instances of program(s)")
	var sFlag = flag.Bool("s", false, "no output, just return 0 if all of the executables are found, or 1 if some were found")
	flag.Parse()

	programs := flag.Args()
	if len(programs) == 0 {
		flag.Usage()
		os.Exit(EXIT_INVALID_ARGS)
	}

	path := os.Getenv("PATH")
	if path == "" {
		fmt.Println("PATH environment variable is empty")
		os.Exit(EXIT_PATH_EMPTY)
	}
	pathSplit := filepath.SplitList(path)

	m := make(map[string]bool)
	var programPaths []string

	// For each program requested, search all paths
	for _, v := range programs {
		found := false

		for _, dir := range pathSplit {
			result := isThere(v, dir)
			if result != "" {
				found = true
				programPaths = append(programPaths, result)

				if !*aFlag {
					break
				}
			}
		}
		m[v] = found
	}

	if *sFlag {
		for _, found := range m {
			if !found {
				os.Exit(EXIT_NOT_ALL_FOUND)
			}
		}

		os.Exit(EXIT_SUCCESS)
	}

	for _, v := range programPaths {
		fmt.Println(v)
	}

	os.Exit(EXIT_SUCCESS)
}

func isThere(file string, path string) string {
	fullPath := filepath.Join(path, file)

	fileInfo, err := os.Stat(fullPath)
	if err != nil {
		return ""
	}

	mode := fileInfo.Mode()
	if !mode.IsRegular() {
		return ""
	}

	if mode&0111 != 0 {
		return fullPath
	}

	return ""
}

func printUsage() {
	fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s [options] program1 [program2 ...]\n", os.Args[0])
	fmt.Println("\nOptions:")
	flag.PrintDefaults()
}

package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

const (
	EXIT_SUCCESS    = 0
	EXIT_FAILURE    = 1 // Ugh, lousy name because we exit for different reasonse but naitve has all exits same code
	EXIT_NONE_FOUND = 2 // OpenBSD
)

func main() {
	flag.Usage = printUsage

	//TODO: how to identify a bad flag; that seems to be a reason for naitve to show usage
	var aFlag = flag.Bool("a", false, "list all instances of program(s)")
	var sFlag = flag.Bool("s", false, "no output, just return 0 if all of the executables are found, or 1 if some were found")
	flag.Parse()

	programs := flag.Args()
	if len(programs) == 0 {
		flag.Usage()
		os.Exit(EXIT_FAILURE)
	}

	path := os.Getenv("PATH")
	if path == "" {
		os.Exit(EXIT_FAILURE)
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

		if runtime.GOOS == "openbsd" && !found {
			programPaths = append(programPaths, "which: "+v+": Command not found.")
		}
	}

	allFound := allFound(m)
	if *sFlag && allFound {
		os.Exit(EXIT_SUCCESS)
	} else if *sFlag && !allFound {
		os.Exit(EXIT_FAILURE)
	}

	for _, v := range programPaths {
		fmt.Println(v)
	}

	// Special results first
	if runtime.GOOS == "openbsd" && noneFound(m) {
		os.Exit(EXIT_NONE_FOUND)
	}

	if allFound {
		os.Exit(EXIT_SUCCESS)
	} else {
		os.Exit(EXIT_FAILURE)
	}

}

func allFound(m map[string]bool) bool {
	for _, found := range m {
		if !found {
			return false
		}
	}
	return true
}

// So far, this is for OpenBSD which returns a 2 if no names were resolved
func noneFound(m map[string]bool) bool {
	for _, found := range m {
		if found {
			return false
		}
	}
	return true
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
	if runtime.GOOS == "darwin" {
		fmt.Println("usage: which [-as] program ...")
	} else if runtime.GOOS == "openbsd" {
		fmt.Println("usage: which [-a] name ...")
	}
}

package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

const EXIT_SUCCESS = 0
const EXIT_FAILURE = 1
const EXIT_NONE_FOUND = 2   // OpenBSD
const EXIT_INVALID_ARGS = 2 // Ubuntu

var errOutput = new(bytes.Buffer)

func main() {
	flag.CommandLine.Init("", flag.ContinueOnError)
	flag.CommandLine.SetOutput(errOutput)

	flag.Usage = printFlagUsage

	var aFlag = flag.Bool("a", false, "list all instances of program(s)")
	var sFlag = flag.Bool("s", false, "no output, just return 0 if all of the executables are found, or 1 if some were found")
	flag.Parse()

	if errOutput.Len() > 0 {
		printUsage()
		return
	}

	programs := flag.Args()
	if len(programs) == 0 {
		if runtime.GOOS != "linux" {
			printUsage()
		}

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

func printFlagUsage() {
	split := strings.Split(errOutput.String(), ":")
	if len(split) != 2 {
		fmt.Println("Invalid error output")
		os.Exit(EXIT_INVALID_ARGS)
	}

	if runtime.GOOS == "darwin" {
		splitDash := strings.Split(split[1], "-")
		if len(splitDash) != 2 {
			fmt.Println("Invalid flag format")
			os.Exit(EXIT_FAILURE)
		}
		fmt.Printf("/usr/bin/which: illegal option -- %v", splitDash[1])
		printUsage()

		os.Exit(EXIT_FAILURE)
	} else if runtime.GOOS == "linux" {
		fmt.Printf("Illegal option%v", split[1])
		printUsage()

		os.Exit(EXIT_INVALID_ARGS)
	} else if runtime.GOOS == "openbsd" {
		// This looks like this: "flag provided but not defined: -z"
		splitOutput := strings.Split(errOutput.String(), "-")
		if len(splitOutput) != 2 {
			fmt.Println("Invalid error output")
			os.Exit(EXIT_FAILURE)
		}

		fmt.Printf("which: unknown option -- %v", splitOutput[1])
		printUsage()

		os.Exit(EXIT_FAILURE)
	}
}

func printUsage() {
	// OpenBSD and Ubuntu differ from OpenBSD, so they are the default case.
	if runtime.GOOS == "openbsd" {
		fmt.Println("usage: which [-a] name ...")
	} else if runtime.GOOS == "linux" {
		fmt.Println("Usage: " + os.Args[0] + " [-as] args ...")
	} else {
		fmt.Println("usage: which [-as] program ...")
	}
}

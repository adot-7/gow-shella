package main

import (
	"fmt"
	"io/fs"
	"os"
	"slices"
	"strings"
)

func handleType(builtinCommands []string, argument string) {
	paths := strings.SplitSeq(os.Getenv("PATH"), string(os.PathListSeparator))
	// paths := strings.Split(os.Getenv("PATH"), string(os.PathListSeparator)) LESS EFFICIENT, as it populates a whole slice, whereas the above
	// function returns an iter.Seq[string] iterator
	// fmt.Printf("%v", paths)
	if slices.Contains(builtinCommands, argument) {
		fmt.Printf("%s is a shell builtin\n", argument)
		return
	}
	for path := range paths {
		directory, err := os.OpenFile(path, os.O_RDONLY, fs.ModeDir)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			panic(err)
		}
		defer directory.Close()
		files, _ := directory.Readdir(-1)
		for _, file := range files {
			if file.Name() == argument && file.Mode()&0100 != 0 {
				fmt.Printf("%s is %s/%s\n", argument, directory.Name(), argument)
				// found = true
				return
			}
		}
	}
	fmt.Printf("%s: not found\n", argument)

}

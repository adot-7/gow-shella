package main

import (
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"slices"
	"strings"
)

func handleType(builtinCommands []string, argument string) {
	if slices.Contains(builtinCommands, argument) {
		fmt.Printf("%s is a shell builtin\n", argument)
		return
	}
	fileinfo, directory := isExecutable(argument)
	if fileinfo != nil {
		fmt.Printf("%s is %s/%s\n", argument, directory, fileinfo.Name())
		return
	}
	fmt.Printf("%s: not found\n", argument)
}

func executeCommand(command string, argument string) {
	argumentSlice := tokenize(argument)
	fileinfo, _ := isExecutable(command)
	if fileinfo == nil {
		fmt.Printf("%s: command not found\n", command)
		return
	}
	cmd := exec.Command(command, argumentSlice...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		fmt.Println(err)
	}
}

func tokenize(argument string) []string {
	argument = strings.TrimSpace(argument)
	var tokens []string
	var current strings.Builder
	isQuotes := false
	isbackslash := false
	var quote rune

	for _, ch := range argument {
		if isbackslash {
			current.WriteRune(ch)
			isbackslash = !isbackslash
			continue
		}
		switch {
		case ch == '\\':
			isbackslash = !isbackslash
		case ch == '"' || ch == '\'':
			if ch == quote || quote == 0 {
				isQuotes = !isQuotes
				quote = ch
			} else {
				current.WriteRune(ch)
			}

		case ch == ' ' && !isQuotes: //if its a space and we are NOT in quotes
			if current.Len() > 0 {
				tokens = append(tokens, current.String())
				current.Reset()
			}
		default:
			current.WriteRune(ch) //simply add to the current buffer
		}
	}
	if current.Len() > 0 {
		tokens = append(tokens, current.String()) //leftover added
	}
	return tokens
}

func isExecutable(command string) (fs.FileInfo, string) {
	paths := strings.SplitSeq(os.Getenv("PATH"), string(os.PathListSeparator))
	// paths := strings.Split(os.Getenv("PATH"), string(os.PathListSeparator)) LESS EFFICIENT, as it populates a whole slice, whereas the above
	// function returns an iter.Seq[string] iterator
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
			if file.Name() == command && file.Mode()&0100 != 0 {
				return file, directory.Name()
			}
		}
	}
	return nil, ""
}

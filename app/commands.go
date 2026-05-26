package main

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"slices"
	"strings"
)

func handleType(builtinCommands []string, argumentSlice []string, outputWriter io.Writer) {
	for _, argument := range argumentSlice {
		if slices.Contains(builtinCommands, argument) {
			fmt.Fprintf(outputWriter, "%s is a shell builtin\n", argument)
			continue
		}
		fileinfo, directory := isExecutable(argument)
		if fileinfo != nil {
			fmt.Fprintf(outputWriter, "%s is %s/%s\n", argument, directory, fileinfo.Name())
			continue
		}
		fmt.Fprintf(outputWriter, "%s: not found\n", argument)
	}
}
func handleExit(outputWriter io.Writer, message string) {
	fmt.Fprintln(outputWriter, message)
	os.Exit(0)
}
func handleEcho(argumentSlice []string, outputWriter io.Writer) {
	result := strings.Join(argumentSlice, " ")
	result = result + "\n"
	outputWriter.Write([]byte(result))
}
func handlePwd(outputWriter io.Writer) {
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return
	}
	fmt.Fprintf(outputWriter, "%s\n", cwd)
}
func handleCd(argument string) {
	dir := returnDir(argument)
	err := os.Chdir(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "cd: %s: No such file or directory\n", argument)
	}
}

func returnDir(argument string) string {
	dir := argument
	if strings.HasPrefix(argument, "~") {
		argument = argument[1:]
		dir = os.Getenv("HOME") + argument
	}
	return dir
}

func parseTokens(inputs []string, builtinCommands []string) {
	isRedirect := false
	isAppend := false
	isErrorWriter := false
	command := inputs[0]
	var arguments []string
	for i, token := range inputs[1:] {
		if token == ">" || token == "1>" || token == ">>" || token == "2>" {
			if i == len(inputs)-3 {
				//means second last token
				//for ["echo", "hello", ">", "file.txt"], len(inputs) is 4,
				// but we iterate from inputs[1:] so position of > becomes 1 which is len(inputs)-3
				isRedirect = true
				if token == ">>" {
					isAppend = true
				} else if token == "2>" {
					isErrorWriter = true
				}
				continue
			}

			fmt.Fprintf(os.Stderr, "Too many arguments after redirect\n")
			return
		}
		if isRedirect && i == len(inputs)-2 { //last iteration
			destination := returnDir(token)
			flags := os.O_WRONLY | os.O_CREATE | os.O_TRUNC
			if isErrorWriter {
				errorWriter, err := os.OpenFile(destination, flags, 0644)
				if err != nil {
					// handleExit(os.Stderr, err.Error())
					fmt.Fprintln(os.Stderr, err.Error())
					return
				}
				defer errorWriter.Close()
				executeCommand(command, arguments, builtinCommands, os.Stdout, errorWriter)
				return
			}
			if isAppend {
				flags = os.O_APPEND | os.O_CREATE | os.O_WRONLY
			}
			outputWriter, err := os.OpenFile(destination, flags, 0644)
			if err != nil {
				// handleExit(os.Stderr, err.Error())
				fmt.Fprintln(os.Stderr, err.Error())
				return
			}
			defer outputWriter.Close()
			executeCommand(command, arguments, builtinCommands, outputWriter, os.Stderr)
			return
		}
		arguments = append(arguments, token)
	}
	executeCommand(command, arguments, builtinCommands, os.Stdout, os.Stderr)

}

func executeCommand(command string, argumentSlice []string, builtinCommands []string, outputWriter io.Writer, errorWriter io.Writer) {
	switch command {
	case "exit":
		handleExit(outputWriter, "")
		return
	case "echo":
		handleEcho(argumentSlice, outputWriter)
		return
	case "type":
		handleType(builtinCommands, argumentSlice, outputWriter)
		return
	case "pwd":
		handlePwd(outputWriter)
		return
	case "cd":
		if len(argumentSlice) > 1 {
			fmt.Fprintln(os.Stderr, "cd: too many arguments")
			return
		}
		handleCd(argumentSlice[0])
		return
	case "":
		return
	}
	fileinfo, _ := isExecutable(command)
	if fileinfo == nil {
		fmt.Fprintf(os.Stderr, "%s: command not found\n", command)
		return
	}
	cmd := exec.Command(command, argumentSlice...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = outputWriter
	cmd.Stderr = errorWriter
	cmd.Run()
	// if err != nil {
	// 	fmt.Fprintln(os.Stderr, err.Error())
	// }
}

func tokenize(argument string) []string {
	argument = strings.TrimSpace(argument)
	var tokens []string
	var current strings.Builder
	isQuotes := false
	isbackslash := false
	var quote rune

	for _, ch := range argument {
		if isbackslash && (!isQuotes || quote == '"') {
			current.WriteRune(ch)
			isbackslash = !isbackslash
			continue
		}
		switch {
		case ch == '\\' && (!isQuotes || quote == '"'):
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

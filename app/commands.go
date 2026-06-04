package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"slices"
	"strings"
	"time"

	"github.com/chzyer/readline"
)

// var completionsDirectory = filepath.Join(returnDir("~"), "/gow-shella/completions")
var completionsMap = make(map[string]string, 0)

func handleType(builtinCommands []string, argumentSlice []string, outputWriter io.Writer) {
	for _, argument := range argumentSlice {
		if slices.Contains(builtinCommands, argument) {
			fmt.Fprintf(outputWriter, "%s is a shell builtin\n", argument)
			continue
		}
		fileName, directory := isExecutable(argument)
		if fileName != "" {
			fmt.Fprintf(outputWriter, "%s is %s/%s\n", argument, directory, fileName)
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
func handleComplete(arguments []string, outputWriter io.Writer, errorWriter io.Writer, prefixCompleter *readline.PrefixCompleter) {
	// if len(arguments) < x {
	// 	fmt.Fprintf(errorWriter, "complete: insufficient arguments\n")
	// 	return
	// }
	// fmt.Println(completionsDirectory)
	flag := arguments[0]
	switch flag {
	case "-p":
		// fmt.Fprintf(errorWriter, "complete: %s: no completion specification\n", arguments[1])
		if len(arguments) < 2 {
			fmt.Fprintf(errorWriter, "complete: insufficient arguments\n")
			return
		}
		checkCompletion(arguments, outputWriter)
	case "-C":
		if len(arguments) < 3 {
			fmt.Fprintf(errorWriter, "complete: insufficient arguments\n")
			return
		}
		registerCompletion(arguments, outputWriter, errorWriter, prefixCompleter)
	}
}

func checkCompletion(arguments []string, outputWriter io.Writer) {
	// fileName
	// dst, err := os.OpenFile(completionsDirectory, os.O_RDWR|os.O_CREATE|os.O_TRUNC, fs.ModeDir)
	// if err!=nil {
	// 	fmt.Fprintf(errorWriter, "complete: completions directory not accessible\n")
	// 	return
	// }
	// files, _ := dst.ReadDir(-1)
	// for _, file := range files{
	// 	if file.IsDir() || file.Name() != {
	// 		continue
	// 	}
	// }
	// dst := filepath.Join(completionsDirectory, arguments[1])
	// f, err := os.Open(dst)
	// if err != nil {
	// 	fmt.Fprintf(outputWriter, "complete: %s: no completion specification\n", arguments[1])
	// 	return
	// }
	// defer f.Close()
	// fmt.Fprintf(outputWriter, "complete -C '%s' %s", dst, arguments[1])
	scriptPath, ok := completionsMap[arguments[1]]
	if !ok {
		fmt.Fprintf(outputWriter, "complete: %s: no completion specification\n", arguments[1])
		return
	}
	fmt.Fprintf(outputWriter, "complete -C '%s' %s\n", scriptPath, arguments[1])
	time.Sleep(99999999)
}
func fetchCompletions(executablePath string) func(string) []string {
	return func(s string) []string {
		cmd := exec.Command("bash", executablePath)
		output, err := cmd.Output()
		if err != nil {
			return []string{""}
		}
		// fmt.Fprintf(os.Stdout, "the output: %s\n", string(output))
		return strings.Fields(string(output))
	}
}

func registerCompletion(arguments []string, outputWriter io.Writer, errorWriter io.Writer, prefixCompleter *readline.PrefixCompleter) {
	/*
		my dumdum went the file way like a normal shell but its more headaceh for codecrafers test so doing in memory

			srcPath := returnDir(arguments[1])
			f, err := os.Open(srcPath)
			if err != nil {
				fmt.Fprintf(errorWriter, "complete: file not accessible\n")
				return
			}
			defer f.Close()
			srcInfo, _ := f.Stat()
			finalDestination := filepath.Join(completionsDirectory, arguments[2])
			dst, err := os.OpenFile(finalDestination, os.O_RDWR|os.O_CREATE|os.O_TRUNC, srcInfo.Mode())
			if err != nil {
				fmt.Fprintf(errorWriter, "complete: completions directory not accessible\n")
				return
			}
			defer dst.Close()
			_, err = io.Copy(dst, f)
			if err != nil {
				fmt.Fprintf(errorWriter, "complete: failed to write completion\n")
				return
			}
	*/
	completionsMap[arguments[2]] = arguments[1]
	// print(prefixCompleter.Tree("    "))
	for _, pc := range prefixCompleter.GetChildren() {
		if strings.TrimSpace(string(pc.GetName())) == arguments[2] {
			// fmt.Fprintf(outputWriter, "%s=%s\n", string(pc.GetName()), arguments[2])
			pc.SetChildren([]readline.PrefixCompleterInterface{
				readline.PcItemDynamic(fetchCompletions(arguments[1]))},
			)
			return
		}
	}
	newCompleters := prefixCompleter.GetChildren()
	newCompleters = append(newCompleters, readline.PcItemDynamic(fetchCompletions(arguments[1])))
	prefixCompleter.SetChildren(newCompleters)
	// print(prefixCompleter.Tree("    "))

}

func returnDir(argument string) string {
	dir := argument
	if strings.HasPrefix(argument, "~") {
		argument = argument[1:]
		dir = os.Getenv("HOME") + argument
	}
	return dir
}

func parseTokens(inputs []string, builtinCommands []string, prefixCompleter *readline.PrefixCompleter) {
	isRedirect := false
	isAppend := false
	isErrorWriter := false
	if len(inputs) == 0 {
		return
	}
	command := inputs[0]
	var arguments []string
	for i, token := range inputs[1:] {
		if token == ">" || token == "1>" || token == ">>" || token == "1>>" || token == "2>" || token == "2>>" {
			if i == len(inputs)-3 {
				//means second last token
				//for ["echo", "hello", ">", "file.txt"], len(inputs) is 4,
				// but we iterate from inputs[1:] so position of > becomes 1 which is len(inputs)-3
				isRedirect = true
				if token == ">>" || token == "1>>" {
					isAppend = true
				} else if token == "2>" || token == "2>>" {
					isErrorWriter = true
					isAppend = true
				}
				continue
			}

			fmt.Fprintf(os.Stderr, "Too many arguments after redirect\n")
			return
		}
		if isRedirect && i == len(inputs)-2 { //last iteration
			destination := returnDir(token)
			flags := os.O_WRONLY | os.O_CREATE | os.O_TRUNC
			if isAppend {
				flags = os.O_APPEND | os.O_CREATE | os.O_WRONLY
			}
			if isErrorWriter {
				errorWriter, err := os.OpenFile(destination, flags, 0644)
				if err != nil {
					// handleExit(os.Stderr, err.Error())
					fmt.Fprintln(os.Stderr, err.Error())
					return
				}
				defer errorWriter.Close()
				executeCommand(command, arguments, builtinCommands, os.Stdout, errorWriter, prefixCompleter)
				return
			}

			outputWriter, err := os.OpenFile(destination, flags, 0644)
			if err != nil {
				// handleExit(os.Stderr, err.Error())
				fmt.Fprintln(os.Stderr, err.Error())
				return
			}
			defer outputWriter.Close()
			executeCommand(command, arguments, builtinCommands, outputWriter, os.Stderr, prefixCompleter)
			return
		}
		arguments = append(arguments, token)
	}
	executeCommand(command, arguments, builtinCommands, os.Stdout, os.Stderr, prefixCompleter)

}

func executeCommand(command string, argumentSlice []string, builtinCommands []string, outputWriter io.Writer, errorWriter io.Writer, prefixCompleter *readline.PrefixCompleter) {
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
	case "complete":
		if len(argumentSlice) == 0 {
			fmt.Fprintln(os.Stderr, "complete: insufficient arguments")
			return
		}
		handleComplete(argumentSlice, outputWriter, errorWriter, prefixCompleter)
		return
	case "":
		return
	}
	// dirSlice := listdirectories("./")("")
	// dirArg := argumentSlice[1]
	// for _, dir := range dirSlice {

	// }
	fileinfo, _ := isExecutable(command)
	if fileinfo == "" {
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

func isExecutable(command string) (string, string) {
	// paths := strings.SplitSeq(os.Getenv("PATH"), string(os.PathListSeparator))
	// paths := strings.Split(os.Getenv("PATH"), string(os.PathListSeparator)) LESS EFFICIENT, as it populates a whole slice, whereas the above
	// function returns an iter.Seq[string] iterator
	// for path := range paths {
	// 	directory, err := os.OpenFile(path, os.O_RDONLY, fs.ModeDir)
	// 	if err != nil {
	// 		if os.IsNotExist(err) {
	// 			continue
	// 		}
	// 		panic(err)
	// 	}
	// 	defer directory.Close()
	// 	files, _ := directory.Readdir(-1)
	// 	for _, file := range files {
	// 		if file.Name() == command && file.Mode()&0100 != 0 {
	// 			return file, directory.Name()
	// 		}
	// 	}
	// }
	dir, ok := executablesMap[command]
	if !ok {
		return "", ""
	}
	return command, dir
}

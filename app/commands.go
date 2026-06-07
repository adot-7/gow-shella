package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/chzyer/readline"
)

// var completionsDirectory = filepath.Join(returnDir("~"), "/gow-shella/completions")
var completionsMap = make(map[string]string, 0)
var jobsCounter = 0
var jobs = make([]job, 0)
var historyAppendIdx = 0

type job struct {
	jobId     int
	recent    string
	processId int
	status    []byte
	trailing  int //0 if running, 1 if done. to truncate the trailing & from done processes. 2 if not to be shown in subsequent jobs command.
	command   string
}

func createJob() job {
	return job{
		status:   bytes.Repeat([]byte(" "), 24),
		recent:   " ",
		trailing: 0,
	}
}

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
		registerCompletion(arguments, prefixCompleter)
	case "-r":
		if len(arguments) < 2 {
			fmt.Fprintf(errorWriter, "complete: insufficient arguments\n")
			return
		}
		removeCompletion(arguments, prefixCompleter)
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
		inputs := make([]string, 3)
		tokenizedInput := tokenize(s)
		// fmt.Println(tokenizedInput)

		inputs[0] = tokenizedInput[0]
		inputs[1] = tokenizedInput[len(tokenizedInput)-1]
		if len(tokenizedInput) > 1 {
			inputs[2] = tokenizedInput[len(tokenizedInput)-2]
		} else {
			inputs[2] = inputs[0]
		}
		// fmt.Println(inputs)
		cmd := exec.Command(executablePath, inputs...)
		env := os.Environ()
		comp_line := fmt.Sprintf("COMP_LINE=%s", s)
		comp_point := strconv.Itoa(len([]byte(s)))
		comp_point = fmt.Sprintf("COMP_POINT=%s", comp_point)
		env = append(env, comp_line)
		env = append(env, comp_point)
		cmd.Env = env
		output, err := cmd.Output()
		if err != nil {
			fmt.Fprintf(os.Stdout, "the output: %s\n", err.Error())
			return []string{""}
		}
		// fmt.Fprintf(os.Stdout, "the output: %v\n", strings.Fields(string(output)))
		return strings.Fields(string(output))
	}
}

func removeCompletion(arguments []string, prefixCompleter *readline.PrefixCompleter) {
	_, ok := completionsMap[arguments[1]]
	if !ok {
		return
	}
	delete(completionsMap, arguments[1])
	for _, pc := range prefixCompleter.GetChildren() {
		if strings.TrimSpace(string(pc.GetName())) == arguments[1] {
			pc.SetChildren([]readline.PrefixCompleterInterface{
				readline.PcItemDynamic(listdirectories("./"))},
			)
			return
		}
	}
}
func registerCompletion(arguments []string, prefixCompleter *readline.PrefixCompleter) {
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
func handleJobTermination(cmd *exec.Cmd, jobId int) {
	err := cmd.Wait()
	if err != nil {
		fmt.Println("damn")
		return
	}
	copy(jobs[jobId-1].status, []byte("Done   "))
	jobs[jobId-1].trailing = 1
}
func startJob(inputs []string, builtinCommands []string, prefixCompleter *readline.PrefixCompleter) {
	if len(inputs) == 1 {
		return
	}

	input := inputs[:len(inputs)-1]
	// go parseTokens(input, builtinCommands, prefixCompleter)
	// fmt.Fprintf(os.Stdout, "[%d] %d", jobs, os.Getpid()) //This could work ig. but not for now, it doesnt reprint
	// inpCombined := strings.Join(input, " ")
	// cmd := exec.Command(input[0], input[1:]...)
	cmd := exec.Command(input[0], input[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Start()
	if err != nil {
		fmt.Fprintf(os.Stdout, "Error: Job cannot be created\n")
		return
	}

	var currJob job
	var jobId int
	reusableId := -1

	for i := range jobs {
		if jobs[i].trailing == 2 {
			reusableId = i
			break
		}
	}
	currJob = createJob()
	currJob.recent = "+"
	currJob.processId = cmd.Process.Pid
	currJob.command = strings.Join(inputs, " ")
	copy(currJob.status, []byte("Running"))

	if reusableId != -1 {
		jobId = jobs[reusableId].jobId
		currJob.jobId = jobId
		jobs[reusableId] = currJob
	} else {
		jobsCounter++
		jobId = jobsCounter
		currJob.jobId = jobId
		jobs = append(jobs, currJob)
	}

	// fmt.Printf("'%s'\n", currJob.status)

	if len(jobs) > 1 {
		// jobs[currJob.jobId-2].recent = "-"
		for i := len(jobs) - 2; i >= 0; i-- {
			if jobs[i].recent == "+" && jobs[i].trailing != 2 {
				jobs[i].recent = "-"
			} else {
				jobs[i].recent = " "
			}
		}
		// for i, _ := range jobs[:currJob.jobId-2] {
		// 	jobs[i].recent = " "
		// }
	}
	fmt.Fprintf(os.Stdout, "[%d] %d\n", currJob.jobId, currJob.processId)
	go handleJobTermination(cmd, currJob.jobId)
}

func handleJobs(outputWriter io.Writer, showRunning bool) {
	for i := range jobs {
		if jobs[i].trailing != 2 {
			// plusRecent := false // true: not updated, false: updated
			switch showRunning {
			case true:
				fmt.Fprintf(outputWriter, "[%d]%s  %s%s\n", jobs[i].jobId, jobs[i].recent, string(jobs[i].status), jobs[i].command[:len(jobs[i].command)-jobs[i].trailing])
			case false:
				if jobs[i].trailing == 1 {
					fmt.Fprintf(outputWriter, "[%d]%s  %s%s\n", jobs[i].jobId, jobs[i].recent, string(jobs[i].status), jobs[i].command[:len(jobs[i].command)-jobs[i].trailing])
				}
			}
			if jobs[i].trailing == 1 {
				jobs[i].trailing = 2
			}
		}
	}
	for i := len(jobs) - 1; i >= 0; i-- {
		if jobs[i].trailing == 0 || jobs[i].trailing == 1 {
			jobs[i].recent = "+"
			for j := i - 1; j >= 0; j-- {
				if jobs[j].trailing == 0 || jobs[j].trailing == 1 {
					jobs[j].recent = "-"
					return
				}
			}
		}
	}
}

func readFromHistory(arguments []string) {
	var historyPath string
	if len(arguments) == 0 {
		fmt.Fprintln(os.Stderr, "history: insufficient arguments")
		return
	}
	if len(arguments) == 1 {
		historyPath = arguments[0]
	} else {
		historyPath = arguments[1]
	}
	f, err := os.ReadFile(historyPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "history: cannot read history file")
		return
	}
	content := string(f)
	lines := strings.Split(content, "\n")
	lines = lines[:len(lines)-1] // account for empty line at end.
	history = append(history, lines...)
}

func writeToHistory(arguments []string) {
	if len(arguments) < 2 {
		fmt.Fprintln(os.Stderr, "history: insufficient arguments")
		return
	}
	historyPath := arguments[1]
	data := strings.Join(history, "\n")
	data = data + "\n"
	err := os.WriteFile(historyPath, []byte(data), 0644)
	if err != nil {
		fmt.Fprintln(os.Stderr, "history: cannot open history file")
		return
	}
}

func appendToHistory(arguments []string) {
	var historyPath string
	if len(arguments) == 0 {
		fmt.Fprintln(os.Stderr, "history: insufficient arguments")
		return
	}
	if len(arguments) == 1 {
		historyPath = arguments[0]
	} else {
		historyPath = arguments[1]
	}
	f, err := os.OpenFile(historyPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Fprintln(os.Stderr, "history: cannot open history file")
		return
	}
	data := strings.Join(history[historyAppendIdx:], "\n")
	data = data + "\n"
	historyAppendIdx = len(history)
	f.Write([]byte(data))
	f.Close()
}

func handleHistory(arguments []string, outputWriter io.Writer) {
	if len(arguments) == 0 {
		for i, item := range history {
			fmt.Fprintf(outputWriter, "\t%d  %s\n", i+1, item)
		}
		return
	}
	switch arguments[0] {
	case "-r":
		readFromHistory(arguments)
		return
	case "-w":
		writeToHistory(arguments)
		return
	case "-a":
		appendToHistory(arguments)
		return
	}
	if len(arguments) == 1 {
		lim, err := strconv.Atoi(arguments[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, "history: %s: numeric argument required\n", arguments[0])
			return
		}
		if lim > len(history) {
			for i, item := range history {
				fmt.Fprintf(outputWriter, "\t%d  %s\n", i+1, item)
			}
			return
		}
		for i := lim; i > 0; i-- {
			fmt.Fprintf(outputWriter, "\t%d  %s\n", len(history)-i+1, history[len(history)-i])
		}
		return
	}

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
				switch token {
				case ">>", "1>>":
					isAppend = true
				case "2>", "2>>":
					isErrorWriter = true
					isAppend = true
				}
				continue
			} else {
				fmt.Fprintf(os.Stderr, "Too many arguments\n")
				return
			}

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
	pipeArgs := slices.Concat([]string{command}, arguments)
	// arguments = arguments[1:]
	pipeCommands := make([][]string, 0)
	j := 0
	// fmt.Printf("%v\n", arguments)
	for i, token := range pipeArgs {
		if token == "|" {
			pipeCommands = append(pipeCommands, pipeArgs[j:i])
			j = i + 1
		}
	}
	if len(pipeCommands) == 0 {
		executeCommand(command, arguments, builtinCommands, os.Stdout, os.Stderr, prefixCompleter)
		return
	}
	pipeCommands = append(pipeCommands, pipeArgs[j:])

	//for each command, set up exec.Command(), store its cmd.stdoutpipe. for the next command, give the stored stdout to the stdin of this command(). if its last, you set up
	// var currStdOut io.ReadCloser
	// var currStdIn io.WriteCloser
	// fmt.Printf("%v: %d\n", pipeCommands, len(pipeCommands))
	type pipeReaderWriter struct {
		r io.Reader
		w io.WriteCloser
	}

	var piper pipeReaderWriter
	var builtinWaitGroup sync.WaitGroup
	// pipers := make(map[int]pipeReaderWriter, 2)
	cmds := make([]*exec.Cmd, len(pipeCommands))

	for i, currCommand := range pipeCommands {
		// cmd := exec.Command("/bin/bash", "-c", currCommand)
		// var cmd *exec.Cmd
		if slices.Contains(builtinCommands, currCommand[0]) && (i == 0 || i == len(pipeCommands)-1) { //first or last pipe command
			// lets only do for the command being the first or the last in the pipeline
			// if on last command, then exec., but how to connect the prev. cmd stdoutpipe to this
			// command which simply takes the input, no pipe. maybe just pass the whole currCommand as input, no need to connect to pipe ig.
			// yeah that might work.
			// if not on last, then exec. but set the output writer to the next command's stdin.
			// but i cannot exec. it just now, i have to somehow start it with other cmds.Start(). I can modify the cmd.Path to something and then check for it.
			// fmt.Printf("adding '%v' to the pipers instead\n", currCommand)
			cmds[i] = exec.Command("echo")
			cmds[i].Path = "builtin" // weird approach ig lol but it might work, giving error
			// cmds[i].Args = currCommand
			// pipers[i] = pipeReaderWriter{r: rdr, w: wrt}
			if i == 0 {
				rdr, wrt := io.Pipe()
				piper = pipeReaderWriter{r: rdr, w: wrt}
			}
		} else if len(currCommand) == 1 {
			cmds[i] = exec.Command(currCommand[0])
		} else {
			cmds[i] = exec.Command(currCommand[0], currCommand[1:]...)
		}
	}

	// _, err := fmt.Printf("%v: %d\n", &cmds[0].Args, len(cmds[0].Args))
	// if err != nil {
	// 	fmt.Printf("damn error: %v\n", err)
	// }
	for i := 0; i <= len(cmds)-2; i++ {
		if cmds[i].Path == "builtin" {
			cmds[i+1].Stdin = piper.r
			continue
		}
		stdoutPipe, err := cmds[i].StdoutPipe()
		if err != nil {
			fmt.Printf("damn error: %v\n", err)
		}
		if cmds[i+1].Path != "builtin" {
			cmds[i+1].Stdin = stdoutPipe
		} else {
			//the builtin will not take any input from the prev. pipe?
			continue
		}
		// TODO: what will happen to the output pipe of the command before the builtin command? since the builtin doesnt take any pipe, only direct input
	}
	// var b bytes.Buffer
	cmds[len(cmds)-1].Stdout = os.Stdout

	for i := len(cmds) - 1; i >= 0; i-- { //starting reverse because:If you start the first command first, it might generate data and write to a pipe whose reading end hasn't opened yet
		if cmds[i].Path == "builtin" {
			builtinWaitGroup.Add(1)
			go func(idx int) {
				defer builtinWaitGroup.Done()
				if idx == 0 { // first pipe
					if !(cmds[idx+1].Path == "builtin") {
						executeCommand(pipeCommands[idx][0], pipeCommands[idx][1:], builtinCommands, piper.w, os.Stderr, prefixCompleter)
						piper.w.Close()

					} else {
						executeCommand(pipeCommands[idx][0], pipeCommands[idx][1:], builtinCommands, io.Discard, os.Stderr, prefixCompleter)
					}
					return
				}
				executeCommand(pipeCommands[idx][0], pipeCommands[idx][1:], builtinCommands, os.Stdout, os.Stderr, prefixCompleter)
			}(i)
			continue
		}
		err := cmds[i].Start()
		if err != nil {
			fmt.Printf("damn error: %v\n", err)
		}
	}
	for _, cmd := range cmds {
		if cmd.Process == nil {
			continue
		}
		err := cmd.Wait()
		if err != nil {
			fmt.Printf("damn error: %v\n", err)
		}
	}
	builtinWaitGroup.Wait()
	// fmt.Printf("buffer: %v\n", b.Len())
	// fmt.Fprintf(os.Stdout, "%s", b.String())
	// time.Sleep(999999999)
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
	case "jobs":
		handleJobs(outputWriter, true)
		return
	case "history":
		if len(argumentSlice) > 2 {
			fmt.Fprintln(os.Stderr, "history: too many arguments")
			return
		}
		handleHistory(argumentSlice, outputWriter)
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

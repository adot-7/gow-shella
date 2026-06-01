package main

import (
	"fmt"
	"io/fs"
	"os"
	"slices"
	"strings"

	"github.com/chzyer/readline"
)

// var completer = readline.NewPrefixCompleter(
//
//	readline.PcItem("echo"),
//	readline.PcItem("exit"),
//
// )
var executablesMap = make(map[string]string)

func filterInput(r rune) (rune, bool) {
	switch r {
	// block CtrlZ feature
	case readline.CharCtrlZ:
		return r, false
	}
	return r, true
}

type RingingAutoCompleter struct {
	handler readline.AutoCompleter
}

func (m *RingingAutoCompleter) Do(line []rune, pos int) ([][]rune, int) {
	newLine, length := m.handler.Do(line, pos)
	slices.SortFunc(newLine, func(a, b []rune) int {
		return strings.Compare(
			string(a), string(b),
		)
	})
	if length == 0 && len(line) > 0 {
		fmt.Print("\a")
	}

	return newLine, length
}

func populateExecutables(completers []readline.PrefixCompleterInterface) []readline.PrefixCompleterInterface {
	paths := strings.SplitSeq(os.Getenv("PATH"), string(os.PathListSeparator))
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
			if !file.IsDir() && file.Mode()&0100 != 0 {
				completers = append(completers, readline.PcItem(file.Name(),
					readline.PcItemDynamic(listdirectories("./"))))
				// fmt.Printf("Adding %s in directory: %s", file.Name(), directory.Name())
				executablesMap[file.Name()] = directory.Name()
			}
		}
	}
	return completers
}
func listdirectories(path string) func(string) []string {
	return func(s string) []string {
		names := make([]string, 0)
		files, _ := os.ReadDir(path)
		for _, f := range files {
			names = append(names, f.Name())
		}
		return names
	}
}

func main() {
	// test()
	loop := true
	builtinCommands := []string{"exit", "echo", "type", "pwd", "cd"}
	var completers []readline.PrefixCompleterInterface
	for _, builtin := range builtinCommands {
		completers = append(completers,
			readline.PcItem(builtin,
				readline.PcItemDynamic(listdirectories("./"))),
		)
		// print(string(&completers[0].Tree()))
		// time.Sleep(500000000)
	}
	completers = populateExecutables(completers)
	// completers = append(completers, readline.PcItemDynamic(listdirectories("./")))
	completer := readline.NewPrefixCompleter(completers...)

	// print(completer.Tree("    "))
	// time.Sleep(500000000)

	// reader := bufio.NewReader(os.Stdin)
	rl, err := readline.NewEx(&readline.Config{
		// Prompt:              "\033[31m$ \033[0m ",
		Prompt:              "$ ",
		HistoryFile:         "/tmp/gowshella.tmp",
		InterruptPrompt:     "^C",
		EOFPrompt:           "exit",
		FuncFilterInputRune: filterInput,
		Stdout:              os.Stdout,
		HistorySearchFold:   true,
		AutoComplete: &RingingAutoCompleter{
			handler: completer,
		},
	})
	if err != nil {
		handleExit(os.Stderr, "Could not create readline instance")
	}
	defer rl.Close()
	rl.CaptureExitSignal()

	for loop {
		// fmt.Print("$ ")

		// acceptableCommands = append(acceptableCommands, "exit")
		// acceptableCommands = append(acceptableCommands, "echo")
		// input, err := reader.ReadString('\n')
		input, err := rl.Readline()
		if err != nil {
			// input, err = reader.ReadString('\n')
			break
		}
		if len(input) == 0 {
			input = ""
		}
		// input = strings.TrimSuffix(input, "\n")
		input = strings.TrimSpace(input)
		// inputs := strings.SplitN(input, " ", 2)
		inputs := tokenize(input)
		parseTokens(inputs, builtinCommands)
		// fmt.Println(inputs[0])
		// command := inputs[0]
		// command = strings.TrimSuffix(command, " ")
		// arguments, err := reader.ReadString('\n')
		// fmt.Printf("arguments: %s\ncommand: %s\n", arguments, command)
		// argumentsSlice := strings.Split(arguments, " ")
		// argumentsSlice = append(argumentsSlice, argumentsSlice...)
		// fmt.Printf("%s\n%s\n", command, arguments)
		// if !slices.Contains(acceptableCommands, command) {
		// 	fmt.Printf("%s: command not found\n", command)
		// 	continue
		// }
		// arguments := ""
		// var arguments []string
		// if len(inputs) > 1 {
		// 	arguments = inputs[1:]
		// }
		// arguments = strings.TrimSpace(arguments)
		// if strings.HasPrefix(command, "\"") {
		// 	command =
		// }
		// executeCommand(command, arguments, builtinCommands, os.Stdout)
	}
}

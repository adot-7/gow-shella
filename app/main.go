package main

import (
	"bufio"
	"os"
	"strings"

	"github.com/chzyer/readline"
)

var completer = readline.NewPrefixCompleter(
	readline.PcItem("echo"),
	readline.PcItem("exit"),
)

func filterInput(r rune) (rune, bool) {
	switch r {
	// block CtrlZ feature
	case readline.CharCtrlZ:
		return r, false
	}
	return r, true
}

func main() {
	loop := true
	builtinCommands := []string{"exit", "echo", "type", "pwd", "cd"}
	reader := bufio.NewReader(os.Stdin)
	rl, err := readline.NewEx(&readline.Config{
		Prompt:              "\033[31m$ \033[0m ",
		HistoryFile:         "/tmp/gowshella.tmp",
		AutoComplete:        completer,
		InterruptPrompt:     "^C",
		EOFPrompt:           "exit",
		FuncFilterInputRune: filterInput,
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
			input, err = reader.ReadString('\n')
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

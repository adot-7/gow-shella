package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	loop := true

	builtinCommands := []string{"exit", "echo", "type", "pwd", "cd"}
	reader := bufio.NewReader(os.Stdin)
	for loop {
		fmt.Print("$ ")
		// acceptableCommands = append(acceptableCommands, "exit")
		// acceptableCommands = append(acceptableCommands, "echo")
		input, err := reader.ReadString('\n')
		if err != nil {
			panic(err)
		}
		input = strings.TrimSuffix(input, "\n")

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

package main

import (
	"bufio"
	"fmt"
	"os"
	"slices"
	"strings"
)

func main() {
	loop := true
	for loop {
		fmt.Print("$ ")
		builtinCommands := []string{"exit", "echo", "type"}
		// acceptableCommands = append(acceptableCommands, "exit")
		// acceptableCommands = append(acceptableCommands, "echo")
		reader := bufio.NewReader(os.Stdin)
		input, err := reader.ReadString('\n')
		if err != nil {
			panic(err)
		}
		input = strings.TrimSuffix(input, "\n")
		inputs := strings.SplitN(input, " ", 2)
		command := inputs[0]
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
		arguments := ""
		if len(inputs) == 2 {
			arguments = inputs[1]
		}

		switch command {
		case "exit":
			loop = false
		case "echo":
			fmt.Printf("%s\n", arguments)
		case "type":
			if slices.Contains(builtinCommands, arguments) {
				fmt.Printf("%s is a shell builtin\n", arguments)
			} else {
				fmt.Printf("%s: not found\n", arguments)
			}
		case "":
			continue
		default:
			fmt.Printf("%s: command not found\n", command)
		}
	}
}

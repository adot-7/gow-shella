package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	loop := true
	builtinCommands := []string{"exit", "echo", "type"}
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
			handleType(builtinCommands, arguments)
		case "":
			continue
		default:
			fmt.Printf("%s: command not found\n", command)
		}
	}
}

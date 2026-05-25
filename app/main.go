package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	loop := true
	builtinCommands := []string{"exit", "echo", "type", "pwd"}
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
		// fmt.Println(inputs[0])
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
		// arguments := ""
		var arguments []string
		if len(inputs) > 1 {
			arguments = inputs[1:]
		}
		// arguments = strings.TrimSpace(arguments)
		// if strings.HasPrefix(command, "\"") {
		// 	command =
		// }
		switch command {
		case "exit":
			loop = false
		case "echo":
			// argumentsSlice := tokenize(arguments)
			// arguments = strings.Join(argumentsSlice, " ")
			result := strings.Join(arguments, " ")
			fmt.Println(result)
		case "type":
			handleType(builtinCommands, arguments)
		case "pwd":
			cwd, err := os.Getwd()
			if err != nil {
				fmt.Println(err)
				continue
			}
			fmt.Printf("%s\n", cwd)
		case "cd":
			if len(arguments) > 1 {
				fmt.Println("cd: too many arguments")
			}
			argument := arguments[0]
			dir := argument
			if argument == "~" {
				dir = os.Getenv("HOME")
			}
			err := os.Chdir(dir)
			if err != nil {
				fmt.Printf("cd: %s: No such file or directory\n", argument)
			}
		case "":
			continue
		default:
			executeCommand(command, arguments)
		}
	}
}

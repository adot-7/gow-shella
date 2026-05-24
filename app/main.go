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
		acceptableCommands := make([]string, 0)
		acceptableCommands = append(acceptableCommands, "exit")
		acceptableCommands = append(acceptableCommands, "echo")
		reader := bufio.NewReader(os.Stdin)
		command, err := reader.ReadString(' ')
		if err != nil {
			panic(err)
		}
		arguments, err := reader.ReadString('\n')
		if err != nil {
			panic(err)
		}
		arguments = strings.TrimSuffix(arguments, "\n")
		command = strings.TrimSuffix(command, " ")
		// fmt.Printf("arguments: %s\ncommand: %s\n", arguments, command)
		// argumentsSlice := strings.Split(arguments, " ")
		// argumentsSlice = append(argumentsSlice, argumentsSlice...)

		if !slices.Contains(acceptableCommands, command) {
			fmt.Printf("%s: command not found\n", command)
		}
		switch command {
		case "exit":
			loop = false
		case "echo":
			fmt.Printf("%s\n", arguments)
		}
	}
}

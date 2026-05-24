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
		command, err := bufio.NewReader(os.Stdin).ReadString('\n')
		if err != nil {
			panic(err)
		}
		command = strings.TrimSuffix(command, "\n")
		if !slices.Contains(acceptableCommands, command) {
			fmt.Printf("%s: command not found\n", command)
		}
		switch command {
		case "exit":
			loop = false
		}
	}
}

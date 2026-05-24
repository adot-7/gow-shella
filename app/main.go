package main

import (
	"bufio"
	"fmt"
	"os"
	"slices"
	"strings"
)

func main() {
	fmt.Print("$ ")
	acceptableCommands := make([]string, 0)
	command, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil {
		panic(err)
	}
	command = strings.TrimSuffix(command, "\n")
	if !slices.Contains(acceptableCommands, command) {
		fmt.Printf("%s: command not found\n", command)
	}
}

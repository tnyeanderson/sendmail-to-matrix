package main

import (
	"bufio"
	"fmt"
	"os"
)

func ask(prompt string) string {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Printf("%s ", prompt)
	scanner.Scan()
	return scanner.Text()
}

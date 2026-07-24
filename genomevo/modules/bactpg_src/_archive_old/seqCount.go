package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
)

func seqCountApp(args []string) {
	if len(args) < 2 {
		fmt.Printf("Lack of input file!\n")
		return
	}

	fileName := args[1]
	file, err := os.Open(fileName)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	count := 0
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, ">") {
			count++
		}
	}

	fmt.Print(count)
}

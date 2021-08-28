package main

import (
	"fmt"
	"os"
	"strconv"
)

func boolToString(v bool) string {
	if v {
		return "true"
	} else {
		return "false"
	}
}

func Itoa(arg int) string

func testInt(name string, expected int, actual int) {
	if expected == actual {
		// printf("[%s]: %d => %d\n", name, expected, actual)
		fmt.Println("[" + name + "]: " + strconv.Itoa(expected) + " => " + strconv.Itoa(actual))
	} else {
		// printf("[%s]: %d expected, but got %d\n", name, expected, actual)
		fmt.Println("[" + name + "]: " + strconv.Itoa(expected) + " expected, but got " + strconv.Itoa(actual))
		os.Exit(1)
	}
}

func testBool(name string, expected bool, actual bool) {
	if expected == actual {
		// printf("[%s]: %hhx => %hhx\n", name, expected, actual)
		fmt.Println("[" + name + "]: " + boolToString(expected) + " => " + boolToString(actual))
	} else {
		// printf("[%s]: %hhx expected, but got %hhx\n", name, expected, actual)
		fmt.Println("[" + name + "]: " + boolToString(expected) + " expected, but got " + boolToString(actual))
		os.Exit(1)
	}
}

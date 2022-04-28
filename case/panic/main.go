package main

import (
	"io"
	"os"
)

func main() {
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		panic(err)
	}
	if string(data) == "bbb" {
		panic("you have found the answer")
	}
	return
}

package main

import (
	"fmt"
	"io"
	"os"
)

func main() {
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		panic(err)
	}
	if string(data) == "bbb" {
		o := &struct {
			a int
			b int
		}{}
		o = nil
		fmt.Printf("%d %d\n", o.a, o.b)
	}
	return
}

package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/vedhavyas/hssh"
)

var addr = flag.String("addr", ":8080", "ip:port")

func main() {
	flag.Parse()
	if *addr == "" {
		fmt.Println("required ip:host for server")
		flag.PrintDefaults()
		os.Exit(1)
	}

	hssh.StartServer(*addr)
}

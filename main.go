package main

import (
	"bytes"
	"io/ioutil"
	"log"
	"os"

	"github.com/alecthomas/repr"
)

//go:generate tool ast.types ast_gen.go main
func main() {
	data, err := ioutil.ReadFile(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}

	p := Parser{}
	repr.Println(p.Parse(bytes.NewReader(data)))
}

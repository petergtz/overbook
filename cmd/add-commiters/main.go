package main

import (
	"fmt"
	"io/ioutil"
)

func main() {

	bytes, e := ioutil.ReadFile("/Users/pego/workspace/go/src/github.com/petergtz/concourse-committers/pipeline.yml")
	if e != nil {
		panic(e)
	}
	output := Manipulate(bytes)

	fmt.Printf(output)
}

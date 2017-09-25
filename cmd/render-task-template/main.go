package main

import (
	"io/ioutil"
	"strconv"
	"strings"

	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	taskConfigPath = kingpin.Arg("task-path", "path to task configuration").Required().String()
)

const templateExt = ".yml.overbook-template"

func main() {
	kingpin.Parse()

	if !strings.HasSuffix(*taskConfigPath, templateExt) {
		kingpin.Fatalf("task file must have extension " + templateExt)
	}

	bytes, e := ioutil.ReadFile(*taskConfigPath)
	if e != nil {
		panic(e)
	}

	d := 10
	for i := 0; i < d; i++ {
		inputs := ""
		for u := 0; u <= i; u++ {
			inputs += "\n- name: input" + strconv.Itoa(u)
		}
		output := strings.Replace(string(bytes), "$INPUTS", inputs, -1)

		ioutil.WriteFile(strings.Replace(*taskConfigPath, templateExt, "", -1)+strconv.Itoa(i+1)+".yml", []byte(output), 0644)
	}

}

package main

import (
	"fmt"
	"io/ioutil"

	"gopkg.in/alecthomas/kingpin.v2"

	"github.com/petergtz/overbook"
)

var (
	pipelineConfigPath = kingpin.Flag("config", "pipeline config yaml path").Required().Short('c').String()
	taskConfigPath     = kingpin.Flag("task-path", "path to task configuration without yml extension").Required().Short('t').String()
	ciResourceName     = kingpin.Flag("ci-resource-name", "the name of the resource containing task scripts").Required().Short('r').String()
)

func main() {
	kingpin.Parse()

	bytes, e := ioutil.ReadFile(*pipelineConfigPath)
	if e != nil {
		panic(e)
	}
	output := concourse.Overbook(bytes, *taskConfigPath, *ciResourceName)

	fmt.Printf(output)
}

package main

import (
	"fmt"
	"io/ioutil"
	"strings"

	"gopkg.in/alecthomas/kingpin.v2"

	"github.com/petergtz/overbook"
)

var (
	pipelineConfigPath  = kingpin.Flag("config", "pipeline config yaml path").Required().Short('c').String()
	taskConfigPath      = kingpin.Flag("task-path", "path to task configuration without yml extension").Required().Short('t').String()
	additionalResources = kingpin.Flag("resource", "the name of an additional resource").Required().Short('r').Strings()
)

func main() {
	kingpin.Parse()

	bytes, e := ioutil.ReadFile(*pipelineConfigPath)
	if e != nil {
		panic(e)
	}
	for _, resource := range *additionalResources {
		if !strings.Contains(resource, "=") {
			kingpin.FatalUsage("resource format for %v not valid", resource)
		}
	}
	output := concourse.Overbook(bytes, *taskConfigPath, *additionalResources)

	fmt.Printf(output)
}

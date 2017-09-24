package concourse

import (
	"fmt"
	"regexp"
	"strconv"

	"github.com/concourse/atc"
	"gopkg.in/yaml.v2"
)

type Payload struct {
	repos                  []string
	previousRepos          []string
	taskFilepathWithoutExt string
	ciResourceName         string
}

func Overbook(pipelineYaml []byte, taskFilepathWithoutExt string, ciResourceName string) string {
	var config atc.Config
	e := yaml.Unmarshal(quoteConcourse(pipelineYaml), &config)
	if e != nil {
		panic(e)
	}

	for u := range config.Jobs {
		context := Payload{
			taskFilepathWithoutExt: taskFilepathWithoutExt,
			ciResourceName:         ciResourceName,
		}
		traverseDo(&config.Jobs[u].Plan, &context)
	}

	output, e := yaml.Marshal(&config)
	if e != nil {
		panic(e)
	}
	return dequoteConcourse(output)
}

func traversAggregate(aggregate *atc.PlanSequence, context *Payload) {
	reposSoFar := make([]string, len(context.repos))
	copy(reposSoFar, context.repos)
	i := 0
	for _, plan := range *aggregate {
		if plan.Aggregate != nil {
			traversAggregate(plan.Aggregate, context)
		}
		if plan.Do != nil {
			traverseDo(plan.Do, context)
		}
		if plan.Get != "" {
			context.repos = append(context.repos, plan.Get)
		}
		if plan.Task != "" {
			if len(reposSoFar) != 0 {
				*aggregate = append((*aggregate)[:i], append([]atc.PlanConfig{task(context, reposSoFar)}, (*aggregate)[i:]...)...)
				i++
			}
		}
		i++
	}
}

func traverseDo(do *atc.PlanSequence, context *Payload) {
	i := 0
	for _, plan := range *do {
		if plan.Aggregate != nil {
			traversAggregate(plan.Aggregate, context)
		}
		if plan.Do != nil {
			traverseDo(plan.Do, context)
		}
		if plan.Get != "" {
			context.repos = append(context.repos, plan.Get)
		}
		if plan.Task != "" {
			if len(context.repos) != len(context.previousRepos) && len(context.repos) != 0 {
				*do = append((*do)[:i], append([]atc.PlanConfig{task(context, context.repos)}, (*do)[i:]...)...)
				context.previousRepos = make([]string, len(context.repos))
				copy(context.previousRepos, context.repos)
				i++
			}
		}
		i++
	}
}

func task(context *Payload, repos []string) atc.PlanConfig {
	return atc.PlanConfig{
		Task:           fmt.Sprintf("Collect inputs from %v", repos),
		TaskConfigPath: fmt.Sprintf("%v%v.yml", context.taskFilepathWithoutExt, len(repos)),
		InputMapping:   inputMapping(repos, context.ciResourceName),
	}
}

func inputMapping(resources []string, ciResource string) map[string]string {
	result := make(map[string]string)
	for i, resource := range resources {
		result["input"+strconv.Itoa(i)] = resource
	}
	result["ci"] = ciResource
	return result
}

var concourseRegex = `\{\{([-\w\p{L}]+)\}\}`

func quoteConcourse(input []byte) []byte {
	re := regexp.MustCompile("(" + concourseRegex + ")")
	return re.ReplaceAll(input, []byte("\"$1\""))
}

func dequoteConcourse(input []byte) string {
	re := regexp.MustCompile("['\"](" + concourseRegex + ")[\"']")
	return re.ReplaceAllString(string(input), "$1")
}

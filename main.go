package main

import (
	"fmt"
	"io/ioutil"
	"regexp"

	"github.com/concourse/atc"
	"gopkg.in/yaml.v2"
)

func main() {

	bytes, e := ioutil.ReadFile("/Users/pego/workspace/go/src/github.com/petergtz/concourse-committers/pipeline.yml")
	if e != nil {
		panic(e)
	}
	output := Manipulate(bytes)

	fmt.Printf(output)
}

func Manipulate(pipelineYaml []byte) string {
	var config atc.Config
	e := yaml.Unmarshal(quoteConcourse(pipelineYaml), &config)
	if e != nil {
		panic(e)
	}

	for u := range config.Jobs {
		var repos, previousRepos []string
		traverseDo(&config.Jobs[u].Plan, &repos, &previousRepos)
	}

	output, e := yaml.Marshal(&config)
	if e != nil {
		panic(e)
	}
	return dequoteConcourse(output)
}

func traversAggregate(aggregate *atc.PlanSequence, repos, previousRepos *([]string)) {
	reposSoFar := make([]string, len(*repos))
	copy(reposSoFar, *repos)
	i := 0
	for _, plan := range *aggregate {
		if plan.Aggregate != nil {
			traversAggregate(plan.Aggregate, repos, previousRepos)
		}
		if plan.Do != nil {
			traverseDo(plan.Do, repos, previousRepos)
		}
		if plan.Get != "" {
			*repos = append(*repos, plan.Get)
		}
		if plan.Task != "" {
			if len(reposSoFar) != 0 {
				*aggregate = append((*aggregate)[:i], append([]atc.PlanConfig{atc.PlanConfig{Task: fmt.Sprintf("Collect committers from %v", reposSoFar)}}, (*aggregate)[i:]...)...)
				i++
			}
		}
		i++
	}
}

func traverseDo(do *atc.PlanSequence, repos, previousRepos *([]string)) {
	i := 0
	for _, plan := range *do {
		if plan.Aggregate != nil {
			traversAggregate(plan.Aggregate, repos, previousRepos)
		}
		if plan.Do != nil {
			traverseDo(plan.Do, repos, previousRepos)
		}
		if plan.Get != "" {
			*repos = append(*repos, plan.Get)
		}
		if plan.Task != "" {
			if len(*repos) != len(*previousRepos) && len(*repos) != 0 {
				*do = append((*do)[:i], append([]atc.PlanConfig{atc.PlanConfig{Task: fmt.Sprintf("Collect committers from %v", *repos)}}, (*do)[i:]...)...)
				*previousRepos = *repos
				i++
			}
		}
		i++
	}
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

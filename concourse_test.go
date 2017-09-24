package concourse_test

import (
	. "github.com/petergtz/overbook"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Concourse", func() {

	It("can manipulate an existing manifest", func() {
		Expect(
			Overbook([]byte(`jobs:
- name: job
  plan:
  - aggregate:
    - get: one
    - get: two
    - task: pre
  - task: Hello
  - task: Hello again
  - get: three
  - task: and again
  - get: four
  - aggregate:
    - get: five
    - task: post`), "generated/repo-aggregation-task"),
		).To(Equal(`groups: []
resources: []
resource_types: []
jobs:
- name: job
  plan:
  - aggregate:
    - get: one
    - get: two
    - task: pre
  - task: Collect inputs from [one two]
    file: generated/repo-aggregation-task2.yml
    input_mapping:
      input0: one
      input1: two
  - task: Hello
  - task: Hello again
  - get: three
  - task: Collect inputs from [one two three]
    file: generated/repo-aggregation-task3.yml
    input_mapping:
      input0: one
      input1: two
      input2: three
  - task: and again
  - get: four
  - aggregate:
    - get: five
    - task: Collect inputs from [one two three four]
      file: generated/repo-aggregation-task4.yml
      input_mapping:
        input0: one
        input1: two
        input2: three
        input3: four
    - task: post
`))
	})
})

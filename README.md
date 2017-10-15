# Overbook

## Introduction

How do you add more activity to a [Concourse](https://concourse.ci)? By overbooking flights.

Overbook is a utility to insert new tasks to an existing Concourse pipeline. The inputs made available to the new tasks are *all* in-resources consumed so far in a job. See "Getting Started" for more details.

## Install

The following instructions assume that you have [Go](https://golang.org/dl/) and [Glide](https://github.com/Masterminds/glide#install) installed on your system and that you [set up a Go workspace properly](https://golang.org/doc/code.html#Workspaces).


    mkdir -p $GOPATH/github.com/petergtz
    cd $GOPATH/github.com/petergtz
    git clone https://github.com/petergtz/overbook.git
    cd overbook
    glide install
    go install github.com/petergtz/overbook/cmd/overbook
    go install github.com/petergtz/overbook/cmd/render-task-template


## Getting Started

### Example Use Case
Let's say every time a Concourse build fails, you want to notify **exactly** the person responsible for the failure, so that unnecessary communication between members of the team can be avoided, and a fix can be pushed as quickly as possible.

One way to do that is to use the [Slack resource](https://github.com/cloudfoundry-community/slack-notification-resource) using _name tags_ so that the person gets notified directly.

Since most of the time changes in pipelines are introduced by git commits, it makes sense to concentrate on committers when it comes to "blaming" the right person.

To trigger the Slack resource on a failure one usually puts it into an `on_failure` block in the job and use [metadata](http://concourse.ci/implementing-resources.html#resource-metadata) similarly to what's [explained](https://github.com/cloudfoundry-community/slack-notification-resource#metadata) in the Slack resource itself.

However, with that approach it's very difficult to tag one or more specific persons in Slack and this is were Overbook can help.

Overbook can augment your pipeline with additional tasks that make all potential breakage candidates available in a file and hence makes them available to an `on_failure` Slack resource.

### Implementing the Example

#### Creating the Committer Aggregation Task
Create a file `aggregate-committers-for-notification.yml.overbook-template` (name can be choosen freely, extension is important) with the following content:

```yaml
platform: linux

image_resource: { type: docker-image, source: { repository: my/docker-repo } }

inputs:
- name: ci
$INPUTS

outputs:
  - name: points-of-contact

run:
  path: ci/scripts/aggregate-committers-for-notification.sh
```

Then,

```sh
render-task-template tasks/aggregate-committers-for-notification.yml.overbook-template
```

generates a task config in different versions where `$INPUTS` gets rendered into different numbers of inputs, e.g. in `aggregate-committers-for-notification3.yml` into:
```yaml
inputs:
- name: ci

- name: input0
- name: input1
- name: input2
```

With that, we can now write `ci/scripts/aggregate-committers-for-notification.sh`. E.g.:
```bash
#! /bin/bash -ex

touch points-of-contact/committers

# Copy all committers from all git input resources into one file:
for input in input*; do
    pushd $input
        if [ ! -e .git ]; then
            popd
            continue
        fi
        now=$(date +%s)
        commit_date=$(git show -s --format=%ct)
        time_since_commit=$((now-commit_date))
        # This is really just a heuristic:
        # commits older than 3 days are not responsible for a broken build
        if [ "$time_since_commit" -gt "259200" ]; then # 3 days = 3*24*60*60 seconds
            popd
            continue
        fi
        echo $(git show -s --format=%ce) >> ../points-of-contact/committers
    popd
done

# for all users do something like this:
sed -e s/my.email@gmail.com/my-slack-user/gI -i points-of-contact/slack-users

# embed slack users in <@...>:
awk '{print "<@" $0 ">"}' points-of-contact/slack-users > points-of-contact/slack-users-with-at
mv points-of-contact/slack-users-with-at points-of-contact/slack-users

# Make slack users available in a single line:
tr '\n' ' ' < points-of-contact/slack-users > points-of-contact/slack-users-single-line
cat points-of-contact/slack-users-single-line
```

Note that the line `for input in input*; do` now automatically takes all inputs made available through the task config.

#### Using the Committer Aggregation Task

With the task from above at hand, we can now augment our pipeline, by running:

```bash
overbook --config pipeline.yml --task-path ci-tasks/tasks/generated/aggregate-committers-for-notification --resource ci=ci-tasks > pipeline-overbooked.yml
```
where `--resource` simply lets us specify additional input-mappings.

`pipeline-overbooked.yml` now has additional tasks which simply make sure all potential blame candidates are available whenever an `on_failure` gets triggered. You can see those tasks in your next `fly set-pipeline`. The only remaining thing to add to our `pipeline.yml` is:

```yaml
notify: &notify
  put: slack
  params:
    text: |
      $TEXT_FILE_CONTENT The Concourse pipeline broke. See:
      $ATC_EXTERNAL_URL/teams/$BUILD_TEAM_NAME/pipelines/$BUILD_PIPELINE_NAME/jobs/$BUILD_JOB_NAME/builds/$BUILD_NAME
    text_file: points-of-contact/slack-users-single-line
```

And on every job a simple:

```yaml
  on_failure: *notify
```

That's it!

## Conclusion

The setup seems pretty complex at first, but there two things to keep in mind:

* Changes to `pipeline.yml` are minimal, and hence the yaml file is not convoluted with "error handling".
* Concoourse does not currently provide a built-in feature to make committers available to a task step.

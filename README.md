# Overbook

## Introduction

How do you add more activity to a [Concourse](https://concourse.ci)? By overbooking flights.

Overbook is a utility to insert new tasks to an existing Concourse pipeline. The inputs made available to the new tasks are *all* in-resources consumed so far in a job. See "Getting Started" for more details.

**Wait, what are you talking about?**

Okay. In simple words, Overbook makes it feasible to systematically notify people on failed jobs in Concourse pipelines. Other approaches require you to copy this notification-on-failure setting all over the place in your pipelines.

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

### Problem Statement
Sometimes it's necessary to run a task in every job of a pipeline, and it takes all resources as input that have been consumed by `get` resources.

While it is possible to add this manually to the pipeline yml, it's a tedious task, involving a lot of copy and paste (with slight changes, because every job might consume a different set of `get` resources), and it litters the pipeline yml with a lot of noise. Furthermore, Concourse task ymls have a fixed set of inputs, which means that for every number of inputs we would need to write its own task yml.

Overbook solves this problem by automating it.

### What Overbook Does
When Overbook runs, it inserts Concourse tasks into jobs. It does this precisely before *every existing task*, with the exception defined below. The inserted task takes all resources consumed so far by the job as input.

The exception to the mechanism explained below, is that Overbook will not insert a task when the set of resources hasn't changed since the last inserted task.

## Example Use Cases

It's not obvious how Overbook can help to solve real-world problems. Before going into such a real-world problem in [How to Slack a Committer of a Faulty Commit](#how_to_slack_a_committer_of_a_faulty_commit), let's look into a simpler use case to get a feel for how Overbook works.

### Hello World

*TBD*

### How to Slack a Committer of a Faulty Commit

#### Problem

Often, in Concourse pipelines, you can see an `on_failure` block in all jobs with the following content:

```yaml
  put: slack
  params:
    text: |
      The Concourse pipeline broke. See:
      $ATC_EXTERNAL_URL/teams/$BUILD_TEAM_NAME/pipelines/$BUILD_PIPELINE_NAME/jobs/$BUILD_JOB_NAME/builds/$BUILD_NAME
```

This notifies the owners of the pipeline that something went wrong.

The problem with this approach is that it doesn't address the actual person who has caused the breakage.

Instead, let's say every time a Concourse build fails, we'd like to notify **exactly** the person responsible for the failure, so that unnecessary communication between members of the team can be avoided, and a fix can be pushed as quickly as possible.

One way to do that is to use Slack's _name tags_ so that the person gets notified directly.

So, roughly what we want to do instead, is:

```yaml
  put: slack
  params:
    text: |
      $TEXT_FILE_CONTENT The Concourse pipeline broke. See:
      $ATC_EXTERNAL_URL/teams/$BUILD_TEAM_NAME/pipelines/$BUILD_PIPELINE_NAME/jobs/$BUILD_JOB_NAME/builds/$BUILD_NAME
    text_file: points-of-contact/slack-users-single-line
```

Where `points-of-contact/slack-users-single-line` contains the Slack users who have potentially caused the broken build. But where does this `points-of-contact/slack-users-single-line` come from?

This is where Overbook comes into play.

#### Implementing the Use Case

##### Creating the Committer Aggregation Task
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

##### Using the Committer Aggregation Task

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

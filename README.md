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

_TBD_
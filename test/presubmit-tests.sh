#!/usr/bin/env bash

# Copyright 2018 The Tekton Authors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# This script runs the presubmit tests; it is started by prow for each PR.
# For convenience, it can also be executed manually.
# Running the script without parameters, or with the --all-tests
# flag, causes all tests to be executed, in the right order.
# Use the flags --build-tests, --unit-tests and --integration-tests
# to run a specific set of tests.

# Markdown linting failures don't show up properly in Gubernator resulting
# in a net-negative contributor experience.
export DISABLE_MD_LINTING=1

source $(dirname $0)/../vendor/github.com/tektoncd/plumbing/scripts/presubmit-tests.sh

function check_go_lint() {
    header "Testing if golint has been done"

    # deadline of 5m, and show all the issues
    golangci-lint -j 1 --color=never run --timeout=5m

    if [[ $? != 0 ]]; then
        results_banner "Go Lint" 1
        exit 1
    fi

    results_banner "Go Lint" 0
}

function check_go_test() {
    header "Testing if go unit test has been done"

    make test-unit

    if [[ $? != 0 ]]; then
        results_banner "Go Unit Test" 1
        exit 1
    fi

    results_banner "Go Unit Test" 0
}

function check_yaml_lint() {
    header "Testing if yamllint has been done"

    make lint-yaml

    if [[ $? != 0 ]]; then
        results_banner "YAML Lint" 1
        exit 1
    fi

    results_banner "YAML Lint" 0
}

function post_build_tests() {
    check_go_lint
    check_go_test
    check_yaml_lint
}

# We use the default build, unit and integration test runners.
main $@

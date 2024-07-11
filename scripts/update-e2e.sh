#!/bin/bash
set -e

# get current develop branch sha
SHA=$(curl https://api.github.com/repos/goplugin/plugin_latest/commits/develop | jq -r '.sha')
echo "Plugin Develop Commit: $SHA"

# update dependencies
go get github.com/goplugin/plugin_latest/integration-tests@$SHA
go mod tidy || echo -e "------\nInitial go mod tidy failed - will update plugin dep and try tidy again\n------"
go get github.com/goplugin/plugin_latest/v2@$SHA
go mod tidy

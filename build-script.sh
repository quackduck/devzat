#!/bin/sh

# This script build Devzat with the correct linker flags to ensure that the
# uname command and similar work.

commit=$(git log --pretty=format:%h --abbrev-commit | head -n 1)
tz=UTC
date=$(TZ=$tz date "+%G-%m-%d")
time=$(TZ=$tz date "+%H:%M")

go build -ldflags "-X 'main.unameCommit=$commit' -X 'main.unameTz=$tz' -X 'main.unameDate=$date' -X 'main.unameTime=$time'"


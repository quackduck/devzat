#!/bin/sh

# This script build Devzat with the correct linker flags to ensure that the
# uname command and similar work.

format_uname () {
	commit=$(git log --pretty=format:%h --abbrev-commit | head -n 1)
	date=$(TZ=UTC date "+Built from commit $commit on the %G-%m-%d at %H:%M (UTC)")
	echo $date
}

go build -ldflags "-X 'main.unameMsg=$(format_uname)'"


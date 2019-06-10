#!/bin/bash

# USAGE
# $ ./release.sh vX.Y.Z

# REQUIREMENTS
# tests must pass

function feedback {
  local TIME=`date +"%Y-%m-%d %H:%M:%S"`
  echo "$TIME $1"
}

VERSION=$1

if test -z $VERSION; then
  echo "missing arg"
  exit 1
fi

feedback "running tests"
go test -race
TEST_RESULT=`echo $?`
if test $TEST_RESULT -ne 0; then
  feedback "tests failed. Exiting"
  exit $TEST_RESULT
fi
feedback "tests pass"

feedback "updating version file"
echo $VERSION > version

for OS in darwin linux; do
  feedback "building hans $VERSION for $OS"
  DIR=bin/$VERSION/$OS
  mkdir -p $DIR
  GOOS=$OS GOARCH=amd64 go build -ldflags "-X hans.Version=$VERSION" -o $DIR/hans ./cmd/hans
done

feedback "adding and pushing new commit"
git add -A
git ci -m $VERSION
git push

feedback "adding and pushing new tag"
git tag $VERSION
git push origin $VERSION

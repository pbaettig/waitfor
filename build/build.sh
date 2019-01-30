#!/bin/bash
set -e

TARGET=${1:-latest}

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
cd $DIR

if [[ "$TARGET" = "release" ]]; then
    git pull --tags >&2

    TAG=$(git describe --tags --abbrev=0)
    if [[ -z $TAG ]]; then
        echo "No tags found." >&2
        exit 1
    else
        COMMIT_ID=$(git rev-list -n 1 $TAG --abbrev-commit)
    fi
    FNAME="$DIR/wfor-$TAG-linux"
elif [[ "$TARGET" = "latest" ]]; then
    TAG="latest"
    COMMIT_ID="$(git log --format="%h" -n 1)"
    FNAME="$DIR/wfor-$COMMIT_ID-linux"
fi



(cd ../cmd/wfor; go test >&2)
sed -E -i "s/Commit([[:space:]]*)= .*/Commit\1= \"$COMMIT_ID\"/g" ../cmd/wfor/version.go
sed -E -i "s/Tag([[:space:]]*)= .*/Tag\1= \"$TAG\"/g" ../cmd/wfor/version.go
sed -E -i "s/BuildDate([[:space:]]*)= .*/BuildDate\1= \"$(date +"%Y-%m-%dT%H:%M:%S")\"/g" ../cmd/wfor/version.go

(cd ../cmd/wfor; go build -o $FNAME)
echo "Successfully built $FNAME"
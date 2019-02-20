#!/bin/bash
# if [[ $# -ne 1 ]]; then
#   echo "usage: ${BASH_SOURCE} [VERSION]"
#   exit 1
# fi

# VERSION=$1

# if [[ $VERSION != "master" ]]; then
#   GH_BASE_URL=https://github.com/heptio/ark/tags/$VERSION
# else
#   GH_BASE_URL=https://github.com/heptio/ark/branches/master
# fi

# svn export $GH_BASE_URL/docs/ $VERSION/ || (echo "Failed to copy docs for version $VERSION" && exit -1)
# svn export --force $GH_BASE_URL/README.md $VERSION/README.md || (echo "Failed to copy README for version $VERSION" && exit -1)


if [[ $# -ne 1 ]]; then
  echo "usage: ${BASH_SOURCE} [VERSION]"
  exit 1
fi

VERSION=$1

# If the version is master, we'll re-use the dir.
if [[ $VERSION != "master" ]]; then
  mkdir $VERSION
fi

# Get the documentation out of the local clone of master and place them into the $VERSION's directory
# --strip-components 1 makes sure we don't end up with docs in $VERSIONS/docs
git archive master docs/ | tar -x --strip-components 1 -C $VERSION
# The README is used as the index
git show master:README.md > $VERSION/README.md
#!/bin/bash
cd "$(dirname "$0")"

if [[ "$(docker images -q alephzero_playground 2> /dev/null)" == "" ]]; then
  echo "TODO: This requires alephzero_playground"
  exit 1
fi

docker run                                          \
  --rm                                              \
  -it                                               \
  -v "${PWD}":/root/go/src/github.com/alephzero/go/ \
  -e "GODEBUG=cgocheck=2"                           \
  -p 12385:12385                                    \
  alephzero_playground

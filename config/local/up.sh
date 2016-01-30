#!/usr/bin/env bash

cd ./shell

function _kill {
  PID=$(ps -xa | grep inotifywait | grep -v grep | head -n 1 | awk '{print $1}' ORS=' ')
  if [[ $PID != "" ]] ; then
    kill $PID
    while [[ $(ps --no-headers $PID) != "" ]] ; do
      sleep 0.1
    done
  fi
}

function _build {
  echo "BUILD"
  go test
}

trap 'exit 0' SIGTERM

echo "OK"
_build
while $(inotifywait -qqr -e MODIFY .); do
  _build
done
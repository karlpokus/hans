#!/bin/bash

trap 'echo "closing db"' SIGINT

while true; do
  echo "speaking"
  sleep 10
done

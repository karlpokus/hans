#!/bin/bash

trap 'echo "cleaning up stuff"; exit 0' SIGINT

while true; do
  echo "speaking"
  sleep 10
done

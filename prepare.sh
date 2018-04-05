#! /bin/bash

# Create the test network
docker network create broker_test_net

set -e

# Clone the emq-docker repo or get it up to date
if [ -d "emq-docker" ]; then
  cd emq-docker
  git pull
else
  git clone https://github.com/emqtt/emq-docker.git emq-docker
  cd emq-docker
fi

# Build the image
docker build . -t emq-docker:latest
cd -

#!/bin/bash

docker build --tag "docker.pkg.github.com/meinto/website-control-graph/website-control-graph:$(cat VERSION)" .
docker push "docker.pkg.github.com/meinto/website-control-graph/website-control-graph:$(cat VERSION)"
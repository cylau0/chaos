#!/bin/sh
export $(cat env | xargs)
docker stack deploy -c docker-compose.yml test_cluster


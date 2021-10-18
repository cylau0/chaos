#!/bin/sh

docker build -t test-server . &&\
(
docker rm -f test-server-1 
docker run -d --rm --name test-server-1 -p 80:80 --env-file env test-server 
)



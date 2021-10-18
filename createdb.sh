#!/bin/sh

mkdir -p ${HOME}/db
docker run -d -v ${HOME}/db:/data/db -p 27017:27017 mongo:4.4


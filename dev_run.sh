#!/usr/bin/env bash
git pull http://joel:20190506@git.0717996.com/Joel/proj_bcbm.git
go build -o bin/bcbm_dev src/server/main.go
cd bin/
nohup ./bcbm_dev &

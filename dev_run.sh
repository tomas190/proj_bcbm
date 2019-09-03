#!/usr/bin/env bash
go build -o bin/bcbm_dev src/server/main.go
cd bin/
nohup ./bcbm_dev &

#!/usr/bin/env bash
go build -o bin/bcbm_dev src/server/main.go
nohup bin/bcbm_dev &

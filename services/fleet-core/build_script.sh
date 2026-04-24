#!/bin/bash
cd /home/abhishek/Desktop/flyfleet/services/fleet-core
go build -v -o fleet-core ./cmd 2>&1 | tee build_v4.log

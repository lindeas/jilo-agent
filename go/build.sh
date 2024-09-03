#!/usr/bin/env bash

###
# Jilo Agent building script
#
# Description: Building script for Jilo Agent
# Author: Yasen Pramatarov
# License: GPLv2
# Project URL: https://lindeas.com/jilo
# Year: 2024
# Version: 0.1
#
# requirements:
# - go
# - upx
###

CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o jilo-agent main.go
upx --best --lzma -o jilo-agent-upx jilo-agent
mv jilo-agent-upx jilo-agent

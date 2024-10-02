#!/usr/bin/env bash

###
# Jilo Agent installation script
#
# Description: Installation script for Jilo Agent
# Author: Yasen Pramatarov
# License: GPLv2
# Project URL: https://lindeas.com/jilo
# Year: 2024
# Version: 0.1
#
###

# systemV
cp jilo-agent.init /etc/init.d/jilo-agent
chmod +x /etc/init.d/jilo-agent
update-rc.d jilo-agent defaults

# systemd
cp jilo-agent.service /lib/systemd/system/jilo-agent.service
systemctl daemon-reload
systemctl enable jilo-agent.service
systemctl start jilo-agent.service

# compatibility with sysV
sudo ln -s /lib/systemd/system/jilo-agent.service /etc/init.d/jilo-agent

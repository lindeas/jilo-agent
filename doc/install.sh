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


# Paths to init and systemd service files
SYSVINIT_SCRIPT="./jilo-agent.init"
SYSTEMD_SERVICE="./jilo-agent.service"
UPSTART_CONF="./jilo-agent.conf"

# Function to install the SysVinit script
install_sysvinit() {

    echo "Detected SysVinit. Installing init script..."
    cp "$SYSVINIT_SCRIPT" /etc/init.d/jilo-agent
    chmod +x /etc/init.d/jilo-agent

    # for Debian/Ubuntu
    if command -v update-rc.d >/dev/null 2>&1; then
        update-rc.d jilo-agent defaults

    # for RedHat/CentOS/Fedora
    elif command -v chkconfig >/dev/null 2>&1; then
        chkconfig --add jilo-agent
    fi

    echo "SysVinit script installed."
}

# Function to install the systemd service file
install_systemd() {

    echo "Detected systemd. Installing systemd service file..."
    cp "$SYSTEMD_SERVICE" /etc/systemd/system/jilo-agent.service
    systemctl daemon-reload
    systemctl enable jilo-agent.service

    # compatibility with sysV
    sudo ln -s /etc/systemd/system/jilo-agent.service /etc/init.d/jilo-agent

    # starting the agent
    systemctl start jilo-agent.service

    echo "Systemd service file installed."
}

# Function to install the Upstart configuration
install_upstart() {

    echo "Detected Upstart. Installing Upstart configuration..."
    cp "$UPSTART_CONF" /etc/init/jilo-agent.conf
    initctl reload-configuration

    echo "Upstart configuration installed."
}

# Detect the init system
if [[ `readlink /proc/1/exe` == */systemd ]]; then
    install_systemd

elif [[ -f /sbin/init && `/sbin/init --version 2>/dev/null` =~ upstart ]]; then
    install_upstart

else
    install_sysvinit

fi

exit 0

#!/usr/bin/env bash

install_dir="/usr/local/bin"
config_dir="/usr/local/etc"

# remove binary executable
sudo rm -rf "$install_dir/chat"
# remove config file
sudo rm -rf "$config_dir/cli-chat-config.json"
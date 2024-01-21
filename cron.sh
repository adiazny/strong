#!/bin/bash

# Add your actual command or script here
# For example, replace "your_command_here" with the command you want to run.
command_to_run="touch /tmp/deleteme/test.txt"

# Set up the cron job to run every 2 minutes
echo "*/2 * * * * $command_to_run" | crontab -


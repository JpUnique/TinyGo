#!/bin/bash
# Set the file descriptor limit
# ulimit -n 200000
# Execute the command provided as arguments to this script
exec "$@"

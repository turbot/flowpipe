#!/bin/bash
set -e

BASE_FOLDER=.devcontainer/scripts

# Start MariaDB
if ! sudo service mariadb status > /dev/null; then
    echo "Starting service MariaDB..."
    sudo service mariadb start

    echo "Running initial setup for MariaDB"
    source $BASE_FOLDER/initialize-mariadb.sh
else
    echo "MariaDB is already running."
fi

# Start PostgreSQL
if ! sudo service postgresql status > /dev/null; then
    echo "Starting service PostgreSQL..."
    sudo service postgresql start

    echo "Running initial setup for PostgreSQL"
    source $BASE_FOLDER/initialize-postgres.sh
else
    echo "PostgreSQL is already running."
fi

# Then run the command provided to docker run, CMD in the Dockerfile, or the override command.
exec "$@"

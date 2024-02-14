#!/bin/bash
set -e

BASE_FOLDER=.devcontainer/scripts

echo "Initialising PostgreSQL..."
sudo -u postgres psql -f $BASE_FOLDER/initialize-postgres.sql > /dev/null

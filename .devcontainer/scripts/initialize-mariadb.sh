#!/bin/bash
set -e

BASE_FOLDER=.devcontainer/scripts

echo "Initialising MariaDB..."
sudo mariadb -u root < $BASE_FOLDER/initialize-mariadb.sql
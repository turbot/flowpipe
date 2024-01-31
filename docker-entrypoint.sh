#!/usr/bin/env bash
# if first arg is anything other than `flowpipe`, assume we want to run flowpipe
# this is for when other commands are passed to the container
if [ "${1:0}" != 'flowpipe' ]; then
    set -- flowpipe "$@"
fi

exec "$@"

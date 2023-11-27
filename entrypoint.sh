#!/bin/sh

export PATH=$PATH:"$(pwd)/bin"

taskapi > /var/log/taskapi.log

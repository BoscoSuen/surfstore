#!/bin/bash
# shellcheck disable=SC2068

# !!!!!! READ BEFORE YOU RUN THE SCRIPT !!!!!!!!!
#    You need to setup your runtime environment variable,
#    such that the script can look up the correct bin directory
#    and run the binaries we just compiled
# 
# If you are using a Mac, open ~/.bash_profile
# If you are using unix/linux machine, open ~/.bashrc
# 
# 1. Append these lines to the bash config file ^^^
#  export GOPATH=<path to starter code>
#  export PATH=$PATH:$GOPATH/bin
#
# 2. Run `source ~/.bash_profile` or `source ~/.bashrc` to make it effective
# 
# 3. Voila

# Clean up the exisitng binaries in current directory
rm -rf ./bin

# Build and install the necessary binaries for scripts to run
cd src/surfstore/
go install ./...
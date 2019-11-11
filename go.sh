#!/bin/bash

repository=${PWD##*/}
echo "publish environment..."
cd ../environment
sh ./publish.sh
cd ../$repository
echo

addr=":10000"
targetos=`uname | tr "[A-Z]" "[a-z]"`
sh ./build.sh $targetos
echo

rm -f ./log/*
log_path=$(pwd)"/log"
db_path=$(pwd)"/data"
echo "run $repository -logpath "$log_path" -addr $addr -dbpath "$db_path
echo
./bin/$repository -logpath $log_path -addr $addr -dbpath $db_path

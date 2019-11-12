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

./bin/$repository 

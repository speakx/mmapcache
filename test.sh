#!/bin/bash

# # # # # # # # # # # # # # # # # # # # # # # # # # # # # # # #
# 测试选项
if [ ! -n "$1" ] ;then
    echo "you need input test target { all | byteio | mmap }."
    exit
else
    echo "the test target is $1"
    echo
fi
target=$1
# # # # # # # # # # # # # # # # # # # # # # # # # # # # # # # #

org=${PWD%/*}
org=${org##*/}
repository=${PWD##*/}
echo "** org:$org"
echo "** repository:$repository"
echo

# 重新造一遍 go mod
sh ./shell/gen-proto.sh
sh ./shell/configure.sh

# byteio test
if [ "$target" == "all" ] || [ "$target" == "byteio" ] ;then
    cd ./src
    go test -v -bench=".*" ./byteio/byteio_test.go ./byteio/byteio.go
    go test -bench=".*" ./byteio/byteio_test.go ./byteio/byteio.go
    go test -v -bench=".*" ./byteio/mem_test.go ./byteio/mem.go
    go test -bench=".*" ./byteio/mem_test.go ./byteio/mem.go
fi

# mmap cache
if [ "$target" == "all" ] || [ "$target" == "mmap" ] ;then
    go get github.com/edsrzf/mmap-go
    cd ./src
    go test -v ./cache/mmapcache_test.go ./cache/mmapcachepool.go ./cache/mmapcache.go ./cache/mmapdata.go
    # go test -bench=".*" ./cache/mmapcache_test.go ./cache/mmapcachepool.go ./cache/mmapcache.go ./cache/mmapdata.go
fi

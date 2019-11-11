#!/bin/bash

# 步骤一
# 安装 protoc
# Mac - 有福了，使用Homebrew就行
# Linux - 稍微麻烦，参考：https://www.jianshu.com/p/04150a3f98b1

# 步骤二
# 更新 plugin
go_path=`go env | grep GOPATH | awk -F'"' '{print $2}'`
echo "GOPATH=$go_path"

git clone https://github.com/grpc/grpc-go.git $go_path/src/google.golang.org/grpc
git clone https://github.com/golang/net.git $go_path/src/golang.org/x/net
git clone https://github.com/golang/text.git $go_path/src/golang.org/x/text
git clone https://github.com/golang/sys.git $go_path/src/golang.org/x/sys
go get -u github.com/golang/protobuf/{proto,protoc-gen-go}
git clone https://github.com/google/go-genproto.git $go_path/src/google.golang.org/genproto

cd $go_path/src/
go install google.golang.org/grpc

# 如果下面这个问题:
# protoc-gen-gogo: program not found or is not executable
# --gogo_out: protoc-gen-gogo: Plugin failed 
# 检查，系统环境变量里是否有GOPATH，可以通过`go env`查看go当前的GOPATH在哪

#!/bin/bash

function gen_proto() {
	protofolder=`ls | grep proto`
    if [ "$protofolder" != "" ] && [ -d "$(pwd)/$protofolder" ] ;then
        cd ./$protofolder
        protos=`ls | grep ".proto"`
        if [ "$protos" != "" ] ;then
            rm -f ./*.pb.go
            protoc --go_out=plugins=grpc:. *.proto
        fi 
        cd ../
    fi
	
    folders=`ls`
	for folder in $folders; do
		if [ -d "$(pwd)/$folder" ] ;then
			cd $folder
				gen_proto
			cd ../
		fi
	done
}

## build proto
echo "build proto..."
export PATH=$PATH:$GOPATH/bin
echo "PATH=$PATH"
echo
gen_proto
echo "gen proto done"
echo
#!/bin/bash

function gen_flatbuffers() {
	schemafolder=`ls | grep schema`
    if [ "$schemafolder" != "" ] && [ -d "$(pwd)/$schemafolder" ] ;then
        cd ./$schemafolder
        fbs=`ls | grep ".fbs"`
        if [ "$fbs" != "" ] ;then
            # rm -f ./*.go
            # rm -r ./
            # protoc --go_out=plugins=grpc:. *.proto

            # 清理生成的fbs的go文件
            rm -f ./*.go
            folders=`ls`
            # 清理之前fbs生成过的目录
            for folder in $folders; do
                if [ -d "$(pwd)/$folder" ] ;then
                    rm -rf ./$folder
                fi
            done

            flatc --go --gen-all --grpc *.fbs
        fi 
        cd ../
    fi
	
    folders=`ls`
	for folder in $folders; do
		if [ -d "$(pwd)/$folder" ] ;then
			cd $folder
				gen_flatbuffers
			cd ../
		fi
	done
}

## build proto
echo "build fbs..."
export PATH=$PATH:$GOPATH/bin
echo "PATH=$PATH"
gen_flatbuffers
echo "gen fbs done"
echo
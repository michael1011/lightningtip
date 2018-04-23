#!/usr/bin/env bash

# Simple bash script to build basic lnd tools for all the platforms
# we support with the golang cross-compiler.
#
# Copyright (c) 2016 Company 0, LLC.
# Use of this source code is governed by the ISC
# license.

# If no tag specified, use date + version otherwise use tag.
if [[ $1x = x ]]; then
    DATE=`date +%Y%m%d`
    VERSION="01"
    TAG=$DATE-$VERSION
else
    TAG=$1
fi

PACKAGE=lightningTip
MAINDIR=$PACKAGE-$TAG
mkdir -p $MAINDIR
cd $MAINDIR

SYS=( "windows-386" "windows-amd64" "linux-386" "linux-amd64" "linux-arm" "linux-arm64")

# GCC cross compiler for the SYS above
# These are necessary for https://github.com/mattn/go-sqlite3
# It is assumed that the machine you are compiling on is running Linux amd64
GCC=( "i686-w64-mingw32-gcc" "x86_64-w64-mingw32-gcc" "gcc" "gcc" "arm-linux-gnueabihf-gcc" "aarch64-linux-gnu-gcc")

# Additional flag to allow cross compiling from 64 to 32 bit on Linux
GCC_LINUX_32BIT="-m32"

# Use the first element of $GOPATH in the case where GOPATH is a list
# (something that is totally allowed).
GPATH=$(echo $GOPATH | cut -f1 -d:)

for index in ${!SYS[@]}; do
    OS=$(echo ${SYS[index]} | cut -f1 -d-)
    ARCH=$(echo ${SYS[index]}  | cut -f2 -d-)

    CC=${GCC[index]}
    CFLAGS=""

    mkdir $PACKAGE-${SYS[index]}-$TAG
    cd $PACKAGE-${SYS[index]}-$TAG

    echo "Building:" $OS $ARCH

    # Add flag to allow cross compilation to 32 bit Linux
    if [[ $OS = "linux" ]]; then
        if [[ $ARCH = "386" ]]; then
            CFLAGS="$GCC_LINUX_32BIT"
        fi
    fi

    env GOOS=$OS GOARCH=$ARCH CGO_ENABLED=1 CC=$CC CFLAGS=$CFLAGS LDFLAGS=$CFLAGS go build github.com/michael1011/lightningtip
    env GOOS=$OS GOARCH=$ARCH CGO_ENABLED=1 CC=$CC CFLAGS=$CFLAGS LDFLAGS=$CFLAGS go build github.com/michael1011/lightningtip/cmd/tipreport

    cd ..

    if [[ $OS = "windows" ]]; then
	zip -r $PACKAGE-${SYS[index]}-$TAG.zip $PACKAGE-${SYS[index]}-$TAG
    else
	tar -cvzf $PACKAGE-${SYS[index]}-$TAG.tar.gz $PACKAGE-${SYS[index]}-$TAG
    fi

    rm -r $PACKAGE-${SYS[index]}-$TAG

done

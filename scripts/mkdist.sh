#!/usr/bin/env bash

# create zip files for distribution on the three major OSes
#
# strictly speaking nothing here should need quoting but that's just my
# environment

go generate
python scripts/checklist.py
unix2dos -n README.md README.txt

version="$(grep -o '".\+"' randomizer/version.go | tr -d '"')"
appname="$(basename "$PWD")"

mkdir -p "dist/$version"

function buildfor() {
	echo "building for $1/$2"
	GOOS=$1 GOARCH=$2 go build
	apack -q "dist/$version/$appname"_$3_"$version.zip" "$appname$4" \
		README.txt checklist/ tracker/
}

buildfor windows amd64 win32 .exe
buildfor darwin amd64 macos64
buildfor linux amd64 linux64

rm README.txt

echo "archives written to dist/$version/"

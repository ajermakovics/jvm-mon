#!/bin/bash

# Convenience script for building the java agent

DIR=`pwd`
MF="$DIR/src/main/resources/MANIFEST.MF"
SRC="$DIR/src/main/java"
MAIN="$SRC/jvmmon/Agent.java"
JAR=jvm-mon-go.jar

echo "Compiling java agent from $SRC"

rm -rf ./build/classes/
rm -rf ./build/libs/

mkdir -p ./build/classes/
mkdir -p ./build/libs/

javac -cp ${SRC} -d build/classes ${MAIN}

cd ./build/classes/
echo "Adding manifest $MF"
jar -cvfm jvm-mon-go.jar ${MF} jvmmon
mv ${JAR} ../libs/

cd ${DIR}
echo "Created agent jar: "
unzip -l ./build/libs/${JAR}

echo "Converting to Go embeddable"
rice embed-go

ls -l ./build/libs/ | grep $JAR
echo "Done"

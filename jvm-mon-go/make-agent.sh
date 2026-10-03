#!/bin/bash

# Convenience script for building the java agent

set -euo pipefail

cd "$(dirname "$0")"
DIR="$(pwd)"
MF="$DIR/src/main/resources/MANIFEST.MF"
SRC="$DIR/src/main/java"
MAIN="$SRC/jvmmon/Agent.java"
JAR=jvm-mon-go.jar

echo "Compiling java agent from $SRC"

rm -rf ./build/classes/
rm -rf ./build/libs/

mkdir -p ./build/classes/
mkdir -p ./build/libs/

# --release 8 when supported (JDK 9+), else fall back to -source/-target
if javac --release 8 -version >/dev/null 2>&1; then
  JAVAC_TARGET="--release 8"
else
  JAVAC_TARGET="-source 8 -target 8"
fi
javac ${JAVAC_TARGET} -cp "${SRC}" -d build/classes "${MAIN}"

cd ./build/classes/
echo "Adding manifest $MF"
jar -cvfm ${JAR} "${MF}" jvmmon
mv ${JAR} ../libs/

cd "${DIR}"
echo "Created agent jar"

ls -l ./build/libs/ | grep $JAR
echo "Done"

[![Tests](https://circleci.com/gh/ajermakovics/jvm-mon.svg?style=shield)](https://circleci.com/gh/ajermakovics/jvm-mon)
![Homebrew](https://img.shields.io/homebrew/v/jvm-mon.svg?colorB=green)
![Release](https://img.shields.io/github/v/release/ajermakovics/jvm-mon?include_prereleases)

# jvm-mon

Console based JVM monitoring - when you just want to SSH into a server and see what's going on.

jvm-top lets you monitor your JVM server applications from the terminal. 

![Screenshot](https://raw.githubusercontent.com/ajermakovics/jvm-mon/pages/site/jvm-mon.png)

# New Version

Release: [1.0-ea1](https://github.com/ajermakovics/jvm-mon/releases/tag/1.0-ea1)
- Rewritten in Go 
- Single executable file
- Can monitor applications on Java 8 and above
- Does not require an existing JDK

How it works:
 - jvm-mon executable comes bundled with a Java agent jar
 - On startup it extracts the agent to a temp directory
 - It attaches to the JVM you want to monitor
 - Loads agent into running JVM to collect metrics
 - Agent and app establish a socket connection to send metrics

# Install

*Requirement*: a JDK8 on the server and `JAVA_HOME` environment variable pointing to it. It won't work with just a JRE.

## MacOS

```
brew install jvm-mon
```

## Linux/MacOS
 1. Download the [release](https://github.com/ajermakovics/jvm-mon/releases) and extract
 2. Set `JAVA_HOME` environment variable: `export JAVA_HOME=/path/to/your/jdk8`
 3. Execute `./bin/jvm-mon` from extracted directory

# Usage

- Select a JVM process and press <kbd>Enter</kbd> to monitor it
- Press <kbd>q</kbd> or <kbd>Ctrl+C</kbd> to exit
- Press <kbd>Del</kbd> or <kbd>Backspace</kbd> to kill a process

# What is available

Currently it shows:
- List of running JVM processes
- Cpu and GC load
- Heap size and usage
- Top threads with cpu usage

# Building from source

To build locally run `./gradlew installDist`.
Then go to `./build/install/jvm-mon/` and run `./bin/jvm-mon`.

To develop you will need `npm` on your machine and then run `./gradlew npmDeps` once to get dependencies.

# How does it work?

jvm-mon is a Kotlin application based on these awesome libraries: 
- [blessed-contrib](https://github.com/yaronn/blessed-contrib) terminal dashboard library in JavaScript
- [J2V8](https://github.com/eclipsesource/J2V8) Java Bindings for V8 JavaScript engine and Node.js
- [jvmtop](https://github.com/patric-r/jvmtop) Java monitoring for the command-line

The way it works is:
 1. The Kotlin app starts a Node.js engine in-process
 2. Node.js loads a script with all the widgets
 3. The script calls back into Kotlin to get metrics



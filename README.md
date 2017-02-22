# jvm-mon

Console based JVM monitoring

jvm-top lets you monitor your Java/Scala/.. server applications from the terminal. 

![sample](https://github.com/ajermakovics/jvm-mon/releases/download/0.1/jvm-mon.png)

# Running

*Requirement*: a JDK8 on the server and `JAVA_HOME` environment variable pointing to it. It won't work with just a JRE.

To run:
 1. Download the [release](https://github.com/ajermakovics/jvm-mon/releases) and extract
 2. Set `JAVA_HOME` environment variable: `export JAVA_HOME=/path/to/your/jdk8`
 3. Execute `./bin/jvm-mon` from extracted directory

Usage:
- Select a JVM process and press `Enter` to monitor it
- Press `q` or `ctrl+c` to exit
- Press `Del` or `Backspace` to kill a process

# What is available

Currently it shows:
- List of running JVM processes
- Cpu and GC load
- Heap size and usage
- Top threads with cpu usage

# Building from source

To build locally run `./gradlew installDist` to install to `./build/install/jvm-mon/`.

To develop you will need `npm` on your machine and then run `./gradlew npmDeps` once to get dependencies.

# How does it work?

jvm-mon is a Java application based on these awesome libraries: 
- [blessed-contrib](https://github.com/yaronn/blessed-contrib) terminal dashboard library in JavaScript
- [J2V8](https://github.com/eclipsesource/J2V8) Java Bindings for V8 JavaScript engine and Node.js
- [jvmtop](https://github.com/patric-r/jvmtop) Java monitoring for the command-line

The way it works is:
 1. The java app starts a Node.js engine in-process
 2. Node.js loads a script with all the widgets
 3. The script calls back into java land to get metrics

# Roadmap/Contributions

Open to suggestions/contributions on what to add. This project was crafted in a one day so there's definitely room for improvment.
Some ideas:
- View process classpath and parameters
- Loaded classes
- Windows build
- Profiler
- View JMX beans


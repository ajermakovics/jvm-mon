# jvm-mon

JVM monitoring from the terminal (cpu, memory, threads).
Single executable written in Go and Java.
Supports monitoring application running on Java 8 and newer.

# Build

Prerequisites:
- Go (at least 1.13)
- https://github.com/Masterminds/glide for dependency management.
- https://github.com/GeertJohan/go.rice for embedded files
- JDK 8+ for building java agent

1. Build java agent: `./gradlew jar`
2. Install Go dependencies: 
```
glide update
glide install
```
3. `go build`

# Usage

To monitor JVMs started with your username:

`./jvm-mon-go`

# How it works

jvm-mon attaches to a running JVM you select and loads an agent.jar into the process.
They communicate via a socket to send/receive JVM metrics in json format.

# Development

Run jvm-mon from Go sources: `./run.sh`
Run a java process: `./agent.sh`

See `log` file for debugging output

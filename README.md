# gobuilder
Advanced builder for complex Go projects

gobuilder is an advanced builder for complex Go projects with many dependencies for use on Linux systems. It's a shell script so you can easily adapt it to your wishes if it doesn't completely suit your needs.

## Features
- Build complete projects including all its dependencies
- Use specific versions or branches of dependencies
- Build for all Go-supported OSes and architectures (can be limited to 64-bit OSes and architectures if your project requires it)
- Do a quick build (dependencies are not updated and not built)

## Help
```
Gobuilder v2.1.0
Usage: ./gobuild -c app_config.cfg [-e GOOS:GOARCH] [-V package@version ] [-lqvw]
       ./gobuild -h
       ./gobuild -s
  -c app_config.cfg   The configuration file to use for this build. Gobuild will look for
                      this file in the following locations, in order:
                        <current_directory>
                        <current_directory>/cfg
                        <home_directory>/.config/gobuilder
                        <gobuilder_directory>/cfg
                        /etc/gobuilder
  -e GOOS:GOARCH      Provide an extra OS/architecture to build for
  -h                  Show this help text
  -l                  Only build for Linux (amd64 architecture)
  -q                  Quick build mode. This skips updating dependencies and will not build them
  -s                  Show supported OS/architecture combinations. Not all combinations
                      will result in a successful build
  -v                  Verbose mode. Shows more output
  -V package@version  Use special version (or branch) for this dependency. Multiple -V parameters are allowed
  -w                  Only build for Windows (amd64 architecture)
```

## Configuration file
The configuration file for Gobuilder contains a number of crucial parameters. In the cfg directory your will find an example file.
These are the fields in the configuration file:
```
# The name of your project
PROJECT_NAME="My project"
# The name of the application you're building with this configuration
APPLICATION_NAME="Hello World"
# The directory where the main.go file can be found
MAIN_PACKAGE_DIR="/go/src/github.com/my_account/helloworld"
# The package name that contains main.go
MAIN_PACKAGE="github.com/my_account/helloworld"
# Where you want your binaries to go. Gobuilder will create <os>/<architecture> subdirectories
# under this directory for each build
BIN_DIR="github.com/my_account/helloworld/bin"
```
Gobuilder will look for this configuration file in the following locations, in order:
- The current directory
- The `cfg` directory under the current directory
- The `.config/gobuilder` directory in your home directory
- The `cfg` directory in the Gobuilder directory
- The `/etc/gobuilder` directory

## OS/architecture combinations
Which OS and architecture combinations will result in a successful build will largely depend on the requirements of your application. For example, building an application
that uses 64-bit datastructures on a 32-bit architecture will fail. 

Different versions of Go may support different OS/architecture combinations.

## What's changed?

version 2.1.0
- This is the first public release
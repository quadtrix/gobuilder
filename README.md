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
Usage: ./gobuild [-e GOOS:GOARCH] [-V package@version ] [-lqvw]
       ./gobuild -h
       ./gobuild -s
  -e GOOS:GOARCH      Provide an extra OS/architecture to build for
  -h                  Show this help text
  -l                  Only build for Linux (amd64 architecture)
  -q                  Quick build mode. This skips updating dependencies and does not build them
  -s                  Show supported OS/architecture combinations. Not all combinations will result in a
                      successful build, this depends on the requirements of your project
  -v                  Verbose mode. Shows more output
  -V package@version  Use special version (or branch) for this dependency. Multiple -V parameters
                      are supported
  -w                  Only build for Windows (amd64 architecture)
```

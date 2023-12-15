# gobuilder
Advanced builder for complex Go projects

gobuilder is an advanced builder for complex Go projects with many dependencies. It can run on any 64-bit OS Go supports.

## Features
- Build complete projects including all its dependencies
- Use specific versions or branches of dependencies
- Build for all Go-supported OSes and architectures
- Do a quick build (dependencies are not updated and not built)

## Help
```
Gobuild v3.0.0
Usage: ./gobuild [-e GOOS:GOARCH] [-V package@version] [-lmnqvw]
       ./gobuild -h
       ./gobuild -s

  -e GOOS:GOARCH      Provide an extra OS/Architecture to build for
  -h                  Show this help
  -l                  Only build for linux:amd64
  -m                  Only build for darwin:amd64 (MacOS)
  -n                  Always use the newest commit of any dependency
  -q                  Do a quick build (does not update or build dependencies)
  -s                  Show Go-supported OS/Architecture combinations. Not all combinations
                      will result in a succesful build
  -v                  Verbose mode. Shows more output
  -V package@version  Use specific version or branch for this dependency. Multiple -V
                      parameters are supported
  -w                  Only build for windows:amd64
```

## Configuration file
The configuration file for Gobuilder contains a number of crucial parameters. In the `cfg` directory your will find the configuration file to build Gobuilder.
The configuration file must always be named `gobuilder.json`. These are the fields in the configuration file:
```
{
    "application_name": "Your application",
    "bin_dir": "/home/your_user/go/src/github.com/your_account/your_application/bin",
    "bin_name": "your_application",
    "main_package": "github.com/your_account/your_application",
    "main_package_dir": "/home/your_user/go/src/github.com/your_account/your_application",
    "project_name": "Your project"
}
```
You are encouraged to use absolute directories in the configuration.
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

## Examples
Show the help
```
gobuild -h
```
Show supported OS/architecture combinations for your version of Go
```
gobuild -s
```
Do a default build of your application for linux:amd64
```
gobuild -l
```
Use version v1.5.0 of the fsnotify package instead of the default 1.6.0 and build for MacOS
```
gobuild -V github.com/fsnotify/fsnotify@v1.5.0 -m
```
Build your application for Android on ARM64 as well as for Linux and Windows
```
gobuild -e android:arm64 -lw
```
Perform a quick build (without dependencies) for Linux only
```
gobuild -ql
```

## What's changed?

version 2.1.0
- This is the first public release
version 3.0.0
- Rewritten in Go
- Modified -l, -w options to not be exclusive anymore
- Added -m option to build for MacOS
- Added -n option to select the newest commits for dependencies
- Removed requirement to provide configuration file on the command line
- Changed configuration file to JSON
- Changed configuration filename to always be the same (see documentation)
- Added configuration option to change the binary name
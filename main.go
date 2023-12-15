package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/quadtrix/basicqueue"
	"github.com/quadtrix/configmanager"
	"github.com/quadtrix/servicelogger"
)

// Build variables
var buildnr string
var builddate string
var builduser string
var osname string
var osversion string

type specialVersion struct {
	_package string
	version  string
}

type gobuild struct {
	buildOS   string
	buildArch string
}

type codevars struct {
	appVersion string
	buildnr    int
	subbuild   string
	builddate  string
	builduser  string
	osname     string
	osversion  string
}

type config struct {
	//configFile      string
	quick           bool
	verbose         bool
	newest          bool
	builds          []gobuild
	specialVersions []specialVersion
	slog            servicelogger.Logger
	cfg             configmanager.Configuration
	applicationName string
	binDir          string
	binName         string
	mainPackage     string
	mainPackageDir  string
	projectName     string
}

type buildRun struct {
	quick        bool
	builds       []gobuild
	failedbuilds []gobuild
}

const (
	appversion string = "3.0.0"
	subbuild   string = ""
)

func getEnvVar(key string) (value string) {
	value = ""
	for _, eitem := range os.Environ() {
		ekey := strings.Split(eitem, "=")[0]
		evalue := strings.Split(eitem, "=")[1]
		if ekey == key {
			return evalue
		}
	}
	return value
}

func showHelp() {
	call := os.Args[0]
	lastslash := strings.LastIndex(call, "/")
	script_path := call[:lastslash-1]
	if script_path == "" {
		script_path = "."
	}
	fmt.Printf("Gobuild v%s.%s%s\n", appversion, buildnr, subbuild)
	fmt.Printf("Usage: %s [-e GOOS:GOARCH] [-V package@version] [-lmnqvw]\n", call)
	fmt.Printf("       %s -h\n", call)
	fmt.Printf("       %s -s\n", call)
	fmt.Println("")
	fmt.Println("  -e GOOS:GOARCH      Provide an extra OS/Architecture to build for")
	fmt.Println("  -h                  Show this help")
	fmt.Println("  -l                  Only build for linux:amd64")
	fmt.Println("  -m                  Only build for darwin:amd64 (MacOS)")
	fmt.Println("  -n                  Always use the newest commit of any dependency")
	fmt.Println("  -q                  Do a quick build (does not update or build dependencies)")
	fmt.Println("  -s                  Show Go-supported OS/Architecture combinations. Not all combinations")
	fmt.Println("                      will result in a succesful build")
	fmt.Println("  -v                  Verbose mode. Shows more output")
	fmt.Println("  -V package@version  Use specific version or branch for this dependency. Multiple -V")
	fmt.Println("                      parameters are supported")
	fmt.Println("  -w                  Only build for windows:amd64")
	os.Exit(1)
}

func (pc *config) saveJson() (err error) {
	savestring := fmt.Sprintf("{\n    \"project_name\": \"%s\",\n    \"application_name\": \"%s\",\n    \"main_package_dir\": \"%s\",\n    \"main_package\": \"%s\",\n    \"bin_dir\": \"%s\",\n    \"bin_name\": \"%s\"\n}", pc.projectName, pc.applicationName, pc.mainPackageDir, pc.mainPackage, pc.binDir, pc.binName)
	return os.WriteFile("./gobuilder.json", []byte(savestring), 0640)
}

func stringPrompt(label string) string {
	var s string
	r := bufio.NewReader(os.Stdin)
	for {
		fmt.Fprint(os.Stderr, label+" ")
		s, _ = r.ReadString('\n')
		if s != "" {
			break
		}
	}
	return strings.TrimSpace(s)
}

func yesNoPrompt(label string) bool {
	var b byte
	r := bufio.NewReader(os.Stdin)
	for {
		fmt.Fprint(os.Stderr, label+" [y/n] ")
		b, _ = r.ReadByte()
		if b != 0 {
			break
		}
	}
	fmt.Fprint(os.Stderr, "\n")
	if b == 'y' || b == 'Y' {
		return true
	} else {
		return false
	}
}

func interactConfigDesign() {
	pc := config{}
	fmt.Println("Answer the following questions to build the configuration file for your application.")
	pc.projectName = stringPrompt("Project name:")
	pc.applicationName = stringPrompt("Application name:")
	if yesNoPrompt("Is the main.go file of your application in the current directory?") {
		pc.mainPackageDir, _ = os.Getwd()
	} else {
		pc.mainPackageDir = stringPrompt("Directory holding main.go of your application:")
	}
	pc.mainPackage = stringPrompt("Name of your module:")
	pc.binDir = stringPrompt("Path to your bin directory (for writing builds to):")
	pc.binName = stringPrompt("Name of your binary (on Windows builds .exe gets added automatically):")
	fmt.Println("This is what your configuration will look like:")
	fmt.Printf("{\n    \"project_name\": \"%s\",\n    \"application_name\": \"%s\",\n    \"main_package_dir\": \"%s\",\n    \"main_package\": \"%s\",\n    \"bin_dir\": \"%s\",\n    \"bin_name\": \"%s\"\n}\n", pc.projectName, pc.applicationName, pc.mainPackageDir, pc.mainPackage, pc.binDir, pc.binName)
	if yesNoPrompt("Do you want to save this configuration to gobuilder.json in the current directory?") {
		pc.saveJson()
	}
}

func showSupported() {
	execObj := exec.Command("go", "tool", "dist", "list")
	var output bytes.Buffer
	execObj.Stdout = &output
	err := execObj.Run()
	if err != nil {
		fmt.Printf("Error: %s\n", err.Error())
	}
	if output.Len() > 0 {
		outlines := strings.Split(output.String(), "\n")
		for _, line := range outlines {
			if line != "" {
				oper := strings.Split(line, "/")[0]
				arch := strings.Split(line, "/")[1]
				fmt.Printf("%s:%s\n", oper, arch)
			}
		}
	}
	os.Exit(2)
}

func (pc *config) load(loglevel servicelogger.LogLevel) error {
	//call := os.Args[0]
	cwd, _ := os.Getwd()
	//	lastslash := strings.LastIndex(call, "/")
	//	script_path := call[:lastslash]
	//	if script_path == "" {
	//		script_path = "."
	//	}
	slog, err := servicelogger.New("nl.quadtrix", os.Stdout.Name(), loglevel, false, "100M", 1)
	if err != nil {
		fmt.Printf("Error: cannot initialize logger: %s\n", err.Error())
		return err
	}
	queue, err := basicqueue.NewJsonQueue(&slog, basicqueue.BQT_BROADCAST, "queue.events", 1000, true, time.Minute)
	if err != nil {
		fmt.Printf("Error: cannot initialize process queueing: %s\n", err.Error())
		return err
	}
	cfg, err := configmanager.New(&slog, queue)
	if err != nil {
		fmt.Printf("Error: cannot initialize config manager: %s\n", err.Error())
		return err
	}
	slog.LogDebug("load", "gobuild", "Looking for gobuilder.json in the following locations:")
	slog.LogDebug("load", "gobuild", fmt.Sprintf("  %s", cwd))
	slog.LogDebug("load", "gobuild", fmt.Sprintf("  %s/.config/gobuilder", getEnvVar("HOME")))
	slog.LogDebug("load", "gobuild", fmt.Sprintf("  %s/cfg", cwd))
	//slog.LogDebug("load", "gobuild", fmt.Sprintf("  %s/cfg", script_path))
	slog.LogDebug("load", "gobuild", "  /etc/gobuilder")
	cfg.SetFilename("gobuilder")
	cfg.SetFiletype(configmanager.CFT_JSON)
	cfg.AddSearchPath(cwd)
	cfg.AddSearchPath(getEnvVar("HOME") + "/.config/gobuilder")
	cfg.AddSearchPath(cwd + "/cfg")
	//cfg.AddSearchPath(script_path + "/cfg")
	cfg.AddSearchPath("/etc/gobuilder")
	err = cfg.ReadConfiguration()
	if err != nil {
		fmt.Printf("Error: unable to load configuration: %s\n", err.Error())
		return err
	}
	pc.cfg = cfg
	pc.slog = slog
	pc.applicationName = cfg.GetString("application_name")
	pc.binDir = cfg.GetString("bin_dir")
	pc.binName = cfg.GetString("bin_name")
	pc.mainPackage = cfg.GetString("main_package")
	pc.mainPackageDir = cfg.GetString("main_package_dir")
	pc.projectName = cfg.GetString("project_name")
	return nil
}

func separateParams(orgparams []string) (separated []string) {
	for _, param := range orgparams {
		if strings.HasPrefix(param, "-") {
			if len(param) > 2 {
				for i, opt := range strings.Split(param, "") {
					if i > 0 {
						separated = append(separated, fmt.Sprintf("-%s", opt))
					}
				}
			} else {
				separated = append(separated, param)
			}
		} else {
			separated = append(separated, param)
		}
	}
	return separated
}

func readParams() (pc config) {
	args := separateParams(os.Args[1:])
	for len(args) > 0 {
		switch args[0] {
		case "-e":
			if len(args) > 1 {
				oa := args[1]
				args = args[1:]
				gb := gobuild{
					buildOS:   strings.Split(oa, ":")[0],
					buildArch: strings.Split(oa, ":")[1],
				}
				pc.builds = append(pc.builds, gb)
			}
		case "-h", "--help":
			showHelp()
		case "-l":
			gb := gobuild{
				buildOS:   "linux",
				buildArch: "amd64",
			}
			pc.builds = append(pc.builds, gb)
		case "-m":
			gb := gobuild{
				buildOS:   "darwin",
				buildArch: "amd64",
			}
			pc.builds = append(pc.builds, gb)
		case "-n":
			pc.newest = true
		case "-q":
			if len(pc.specialVersions) > 0 {
				fmt.Println("-q and -V options cannot be combined")
				os.Exit(6)
			}
			pc.quick = true
		case "-s":
			showSupported()
		case "-v":
			pc.verbose = true
		case "-V":
			if pc.quick {
				fmt.Println("-q and -V options cannot be combined")
				os.Exit(7)
			}
			if len(args) > 1 {
				packver := args[1]
				args = args[1:]
				sv := specialVersion{
					_package: strings.Split(packver, "@")[0],
					version:  strings.Split(packver, "@")[1],
				}
				pc.specialVersions = append(pc.specialVersions, sv)
			} else {
				fmt.Println("Error: The -V parameter needs a package@version")
				showHelp()
			}
		case "-w":
			gb := gobuild{
				buildOS:   "windows",
				buildArch: "amd64",
			}
			pc.builds = append(pc.builds, gb)
		default:
			fmt.Printf("Error: unknown parameter %s\n", args[0])
			showHelp()
		}
		args = args[1:]
	}
	return pc
}

func (pc *config) getCodeVars() (cv codevars, err error) {
	if _, err = os.Stat(pc.mainPackageDir + "/buildnr"); err == nil {
		buff, err := os.ReadFile(pc.mainPackageDir + "/buildnr")
		if err != nil {
			return cv, err
		}
		bnr, err := strconv.Atoi(strings.TrimSpace(string(buff)))
		if err != nil {
			return cv, err
		}
		cv.buildnr = bnr + 1
		os.WriteFile(pc.mainPackageDir+"/buildnr", []byte(fmt.Sprintf("%d", cv.buildnr)), 0640)
	} else {
		pc.slog.LogInfo("getCodeVars", "gobuild", "buildnr not found, assuming 0")
		cv.buildnr = 0
	}
	cv.builddate = time.Now().Format("2006-01-02 15:04:05")
	bu, err := user.Current()
	if err != nil {
		cv.builduser = "anon@"
	} else {
		cv.builduser = bu.Username + "@"
	}
	hostn, err := os.Hostname()
	if err != nil {
		cv.builduser = cv.builduser + "unknown"
	} else {
		cv.builduser = strings.TrimSpace(cv.builduser + hostn)
	}
	if runtime.GOOS == "windows" {
		cmd := exec.Command("systeminfo")
		var output bytes.Buffer
		cmd.Stdout = &output
		err := cmd.Run()
		if err != nil {
			return cv, err
		}
		outlines := strings.Split(output.String(), "\n")
		for _, line := range outlines {
			if strings.HasPrefix(line, "OS Name:") {
				cv.osname = strings.TrimSpace(strings.TrimPrefix(line, "OS Name:"))
			}
			if strings.HasPrefix(line, "OS Version:") {
				cv.osversion = strings.TrimSpace(strings.TrimPrefix(line, "OS Version:"))
			}
		}
	} else {
		cmd := exec.Command("uname", "-sn")
		var output bytes.Buffer
		cmd.Stdout = &output
		err = cmd.Run()
		if err != nil {
			return cv, err
		}
		cv.osname = strings.TrimSuffix(output.String(), "\n")
		cmd = exec.Command("uname", "-r")
		output.Reset()
		cmd.Stdout = &output
		err = cmd.Run()
		if err != nil {
			return cv, err
		}
		cv.osversion = strings.TrimSuffix(output.String(), "\n")
	}

	f, err := os.Open(pc.mainPackageDir + "/main.go")
	if err != nil {
		return cv, err
	}
	defer f.Close()
	pc.slog.LogDebug("getCodeVars", "gobuild", fmt.Sprintf("Opened %s/main.go", pc.mainPackageDir))
	scanner := bufio.NewScanner(f)
	appv_found := false
	subb_found := false
	for scanner.Scan() {
		appv, _ := regexp.MatchString("appversion.*string.*=", scanner.Text())
		subb, _ := regexp.MatchString("subbuild.*string.*=", scanner.Text())
		if appv && !appv_found {
			cv.appVersion = strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(strings.Fields(scanner.Text())[3], "\""), "\""))
			appv_found = true
		}
		if subb && !subb_found {
			cv.subbuild = strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(strings.Fields(scanner.Text())[3], "\""), "\""))
			subb_found = true
		}
	}
	pc.slog.LogDebug("getCodeVars", "gobuild", fmt.Sprintf("appversion = %s", cv.appVersion))
	pc.slog.LogDebug("getCodeVars", "gobuild", fmt.Sprintf("buildnr    = %d", cv.buildnr))
	pc.slog.LogDebug("getCodeVars", "gobuild", fmt.Sprintf("subbuild   = %s", cv.subbuild))
	pc.slog.LogDebug("getCodeVars", "gobuild", fmt.Sprintf("builddate  = %s", cv.builddate))
	pc.slog.LogDebug("getCodeVars", "gobuild", fmt.Sprintf("builduser  = %s", cv.builduser))
	pc.slog.LogDebug("getCodeVars", "gobuild", fmt.Sprintf("osname     = %s", cv.osname))
	pc.slog.LogDebug("getCodeVars", "gobuild", fmt.Sprintf("osversion  = %s", cv.osversion))

	return cv, nil
}

func (pc *config) runGoGet(packageName string, packageVersion string) (err error) {
	var cmd *exec.Cmd
	if packageVersion == "" {
		cmd = exec.Command("go", "get", packageName)
	} else {
		cmd = exec.Command("go", "get", fmt.Sprintf("%s@%s", packageName, packageVersion))
	}
	stderr, _ := cmd.StderrPipe()
	scanner := bufio.NewScanner(stderr)
	err = cmd.Start()
	if err != nil {
		pc.slog.LogError("runGoGet", "gobuild", fmt.Sprintf("go get %s@%s failed to start: %s", packageName, packageVersion, err.Error()))
		return err
	}
	for scanner.Scan() {
		pc.slog.LogDebug("runGoGet", "gobuild", scanner.Text())
	}
	err = cmd.Wait()
	if err != nil {
		pc.slog.LogError("runGoGet", "gobuild", fmt.Sprintf("go get %s@%s failed: %s", packageName, packageVersion, err.Error()))
		return err
	}
	return nil
}

func (pc *config) findVersion(packageName string) (packageVersion string) {
	for _, specver := range pc.specialVersions {
		if specver._package == packageName {
			return specver.version
		}
	}
	return "HEAD"
}

func (pc *config) runBuilds() (br buildRun, err error) {
	br.quick = pc.quick
	br.builds = pc.builds
	if len(pc.specialVersions) > 0 {
		pc.slog.LogInfo("runBuilds", "gobuild", "Specific versions of packages:")
		for _, specver := range pc.specialVersions {
			pc.slog.LogInfo("runBuilds", "gobuild", fmt.Sprintf("  %s: %s", specver._package, specver.version))
		}
	}
	cwd, _ := os.Getwd()
	if cwd != pc.mainPackageDir {
		pc.slog.LogDebug("runBuilds", "gobuild", fmt.Sprintf("Current working directory: %s", cwd))
		pc.slog.LogDebug("runBuilds", "gobuild", fmt.Sprintf("Changing to directory %s", pc.mainPackageDir))
		err = os.Chdir(pc.mainPackageDir)
		if err != nil {
			pc.slog.LogError("runBuilds", "gobuild", fmt.Sprintf("Unable to change to main package directory: %s", err.Error()))
			return br, err
		}
	}
	codeVars, err := pc.getCodeVars()
	if err != nil {
		pc.slog.LogError("runBuilds", "gobuild", fmt.Sprintf("Error determining build flags: %s", err.Error()))
	}
	if !br.quick {
		pc.slog.LogDebug("runBuilds", "gobuild", "Rebuilding go.mod / go.sum")
		if _, err = os.Stat("go.mod"); err == nil {
			pc.slog.LogDebug("runBuilds", "gobuild", "Removing existing go.mod and go.sum")
			err = os.Remove("go.mod")
			if err != nil {
				pc.slog.LogError("runBuilds", "gobuild", fmt.Sprintf("Unable to remove go.mod: %s", err.Error()))
				return br, err
			}
			err = os.Remove("go.sum")
			if err != nil {
				pc.slog.LogError("runBuilds", "gobuild", fmt.Sprintf("Unable to remove go.sum: %s", err.Error()))
				return br, err
			}
		}
		pc.slog.LogDebug("runBuilds", "gobuild", fmt.Sprintf("Running go mod init %s", pc.mainPackage))
		cmd := exec.Command("go", "mod", "init", pc.mainPackage)
		stderr, _ := cmd.StderrPipe()
		scanner := bufio.NewScanner(stderr)
		err = cmd.Start()
		if err != nil {
			pc.slog.LogError("runBuilds", "gobuild", fmt.Sprintf("Failed to start go mod init %s: %s", pc.mainPackage, err.Error()))
			return br, err
		}
		for scanner.Scan() {
			pc.slog.LogDebug("runBuilds", "gobuild", scanner.Text())
		}
		err = cmd.Wait()
		if err != nil {
			pc.slog.LogError("runBuilds", "gobuild", fmt.Sprintf("go mod init %s failed: %s", pc.mainPackage, err.Error()))
			return br, err
		}
		pc.slog.LogInfo("runBuilds", "gobuild", fmt.Sprintf("Updating %s module dependencies...", pc.applicationName))
		for _, specver := range pc.specialVersions {
			pc.slog.LogDebug("runBuilds", "gobuild", fmt.Sprintf("Running go mod edit -require %s@%s", specver._package, specver.version))
			cmd = exec.Command("go", "mod", "edit", "-require", specver._package+"@"+specver.version)
			err = cmd.Run()
			if err != nil {
				pc.slog.LogError("runBuilds", "gobuild", fmt.Sprintf("Unable to run go mod edit -require %s@%s: %s", specver._package, specver.version, err.Error()))
				return br, err
			}
		}
		pc.slog.LogInfo("runBuilds", "gobuild", "Determining dependencies...")
		cmd = exec.Command("go", "mod", "tidy")
		stderr, _ = cmd.StderrPipe()
		scanner = bufio.NewScanner(stderr)
		err = cmd.Start()
		if err != nil {
			pc.slog.LogError("runBuilds", "gobuild", fmt.Sprintf("Unable to run go mod tidy: %s", err.Error()))
			return br, err
		}
		for scanner.Scan() {
			pc.slog.LogInfo("runBuilds", "gobuild", scanner.Text())
		}
		err = cmd.Wait()
		if err != nil {
			pc.slog.LogError("runBuilds", "gobuild", fmt.Sprintf("go mod tidy failed: %s", err.Error()))
			return br, err
		}
		// run go get for each package in go.mod
		gomod, err := os.ReadFile(pc.mainPackageDir + "/go.mod")
		if err != nil {
			pc.slog.LogError("runBuilds", "gobuild", fmt.Sprintf("Failed to open go.mod: %s", err.Error()))
			return br, err
		}
		var gomodpackages []string
		gomodlines := strings.Split(string(gomod), "\n")
		for _, gomodline := range gomodlines {
			if strings.HasPrefix(gomodline, "module") ||
				strings.HasPrefix(gomodline, "go") ||
				strings.HasPrefix(gomodline, "require") ||
				strings.HasPrefix(gomodline, "}") ||
				strings.HasPrefix(gomodline, ")") ||
				gomodline == "" {
				continue
			}
			gomodpackage := strings.Split(strings.TrimSpace(gomodline), " ")[0]
			gomodpackages = append(gomodpackages, gomodpackage)
			pc.slog.LogDebug("runBuilds", "gobuild", fmt.Sprintf("Found required package: %s", gomodpackage))
		}
		for _, gmpac := range gomodpackages {
			gmver := ""
			if pc.newest {
				gmver = pc.findVersion(gmpac)
			}
			err = pc.runGoGet(gmpac, gmver)
			if err != nil {
				pc.slog.LogError("runBuilds", "gobuild", "Dependency update failed, see above for the cause")
				return br, err
			}
		}
	}
	var cmd *exec.Cmd
	if _, err = os.Stat(pc.binDir); err == nil {
		pc.slog.LogDebug("runBuilds", "gobuild", fmt.Sprintf("bin_dir %s already exists", pc.binDir))
	} else {
		pc.slog.LogDebug("runBuilds", "gobuild", fmt.Sprintf("Creating bin_dir %s", pc.binDir))
		err = os.MkdirAll(pc.binDir, 0750)
		if err != nil {
			pc.slog.LogError("runBuilds", "gobuild", fmt.Sprintf("Unable to create bin_dir (%s): %s", pc.binDir, err.Error()))
			return br, err
		}
	}
	pc.slog.LogInfo("runBuilds", "gobuild", "Starting build process. Building for the following targets:")
	for _, build := range br.builds {
		pc.slog.LogInfo("runBuilds", "gobuild", fmt.Sprintf("  %s:%s", build.buildOS, build.buildArch))
	}
	for _, build := range br.builds {
		pc.slog.LogInfo("runBuilds", "gobuild", fmt.Sprintf("Building %s %s.%d%s - Target: %s:%s", pc.applicationName, codeVars.appVersion, codeVars.buildnr, codeVars.subbuild, build.buildOS, build.buildArch))
		outputdir := pc.binDir + "/" + build.buildOS + "/" + build.buildArch
		if _, err = os.Stat(outputdir); err != nil {
			pc.slog.LogInfo("runBuilds", "gobuild", fmt.Sprintf("Creating output directory %s", outputdir))
			err = os.MkdirAll(outputdir, 0750)
			if err != nil {
				pc.slog.LogError("runBuilds", "gobuild", fmt.Sprintf("Unable to create output directory %s: %s", outputdir, err.Error()))
				return br, err
			}
		}
		if pc.verbose {
			if pc.quick {
				cmd = exec.Command("env",
					fmt.Sprintf("GOOS=%s", build.buildOS),
					fmt.Sprintf("GOARCH=%s", build.buildArch),
					"go", "build", "-v", "-ldflags",
					fmt.Sprintf("-X 'main.buildnr=%d' -X 'main.builddate=%s' -X 'main.builduser=%s' -X 'main.osname=%s' -X 'main.osversion=%s'", codeVars.buildnr, codeVars.builddate, codeVars.builduser, codeVars.osname, codeVars.osversion),
					"-o",
					fmt.Sprintf("%s/%s/%s/%s", pc.binDir, build.buildOS, build.buildArch, pc.binName))
			} else {
				cmd = exec.Command("env",
					fmt.Sprintf("GOOS=%s", build.buildOS),
					fmt.Sprintf("GOARCH=%s", build.buildArch),
					"go", "build", "-v", "-a", "-ldflags",
					fmt.Sprintf("-X 'main.buildnr=%d' -X 'main.builddate=%s' -X 'main.builduser=%s' -X 'main.osname=%s' -X 'main.osversion=%s'", codeVars.buildnr, codeVars.builddate, codeVars.builduser, codeVars.osname, codeVars.osversion),
					"-o",
					fmt.Sprintf("%s/%s/%s/%s", pc.binDir, build.buildOS, build.buildArch, pc.binName))
			}
		} else {
			if pc.quick {
				cmd = exec.Command("env",
					fmt.Sprintf("GOOS=%s", build.buildOS),
					fmt.Sprintf("GOARCH=%s", build.buildArch),
					"go", "build", "-v", "-ldflags",
					fmt.Sprintf("-X 'main.buildnr=%d' -X 'main.builddate=%s' -X 'main.builduser=%s' -X 'main.osname=%s' -X 'main.osversion=%s'", codeVars.buildnr, codeVars.builddate, codeVars.builduser, codeVars.osname, codeVars.osversion),
					"-o",
					fmt.Sprintf("%s/%s/%s/%s", pc.binDir, build.buildOS, build.buildArch, pc.binName))
			} else {
				cmd = exec.Command("env",
					fmt.Sprintf("GOOS=%s", build.buildOS),
					fmt.Sprintf("GOARCH=%s", build.buildArch),
					"go", "build", "-v", "-a", "-ldflags",
					fmt.Sprintf("-X 'main.buildnr=%d' -X 'main.builddate=%s' -X 'main.builduser=%s' -X 'main.osname=%s' -X 'main.osversion=%s'", codeVars.buildnr, codeVars.builddate, codeVars.builduser, codeVars.osname, codeVars.osversion),
					"-o",
					fmt.Sprintf("%s/%s/%s/%s", pc.binDir, build.buildOS, build.buildArch, pc.binName))
			}
		}
		argstr := ""
		for _, arg := range cmd.Args {
			argstr = argstr + " " + arg
		}
		pc.slog.LogDebug("runBuilds", "gobuild", fmt.Sprintf("%s %s", cmd.Path, argstr))
		stderr, _ := cmd.StderrPipe()
		err = cmd.Start()
		if err != nil {
			pc.slog.LogError("runBuilds", "gobuild", fmt.Sprintf("Unable to start build: %s", err.Error()))
			return br, err
		}
		scanner := bufio.NewScanner(stderr)
		scanner.Split(bufio.ScanWords)
		for scanner.Scan() {
			pc.slog.LogDebug("runBuilds", "gobuild", scanner.Text())
		}
		err = cmd.Wait()
		if err != nil {
			pc.slog.LogError("runBuilds", "gobuild", fmt.Sprintf("Build of %s:%s failed: %s", build.buildOS, build.buildArch, err.Error()))
			br.failedbuilds = append(br.failedbuilds, build)
		}
	}
	return br, nil
}

func main() {
	// Read command line variables
	pConfig := readParams()
	var loglevel servicelogger.LogLevel
	if pConfig.verbose {
		loglevel = servicelogger.LL_DEBUG
	} else {
		loglevel = servicelogger.LL_INFO
	}
	err := pConfig.load(loglevel)
	if err != nil {
		if yesNoPrompt("Configuration file gobuilder.json not found. Do you want to build one now?") {
			interactConfigDesign()
			err = pConfig.load(loglevel)
			if err != nil {
				fmt.Printf("Error loading configuration: %s\n", err.Error())
				os.Exit(5)
			}
		}
	}
	if len(pConfig.builds) == 0 {
		pConfig.slog.LogError("main", "gobuild", "At least one build target should be specified. Use options -e, -l, -m and/or -w")
		showHelp()
	}
	pConfig.slog.LogInfo("main", "gobuild", fmt.Sprintf("Gobuilder v%s.%s%s", appversion, buildnr, subbuild))
	pConfig.slog.LogDebug("main", "gobuild", fmt.Sprintf("Built on %s by %s on %s (%s)", builddate, builduser, osname, osversion))
	pConfig.slog.LogDebug("main", "gobuild", fmt.Sprintf("Reading configuration from %s", pConfig.cfg.GetRealFilename()))
	buildResult, err := pConfig.runBuilds()
	if err != nil {
		pConfig.slog.LogError("main", "gobuild", fmt.Sprintf("Error during builds: %s", err.Error()))
		os.Exit(4)
	}
	pConfig.slog.LogInfo("main", "gobuild", fmt.Sprintf("Builds complete, %d build(s) failed", len(buildResult.failedbuilds)))
	if len(buildResult.failedbuilds) > 0 {
		pConfig.slog.LogInfo("main", "gobuild", "The following builds have failed:")
		for _, failedbuild := range buildResult.failedbuilds {
			pConfig.slog.LogInfo("main", "gobuild", fmt.Sprintf("%s:%s", failedbuild.buildOS, failedbuild.buildArch))
		}
	}
}

package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"syscall"
)

const (
	domainName = "example.com"
	gitRoot = "github.com/ExampleUser"
	portNumber = "8000"
)

// Daemonise the process so it is runnable from a service manager such as BSD init or systemd
func daemonise() {
	// Drop privileges by switching to nobody user and group
	if _, _, err := syscall.Syscall(syscall.SYS_SETGID, 65534, 0, 0); err != 0 {
		os.Exit(1)
	}
	if _, _, err := syscall.Syscall(syscall.SYS_SETUID, 65534, 0, 0); err != 0 {
		os.Exit(1)
	}

	// Do first fork
	pid, _, _ := syscall.Syscall(syscall.SYS_FORK, 0, 0, 0)

	// Exit in parent process
	switch pid {
	case 0:
		// Child process, carry on
		break
	default:
		// Parent process, exit cleanly
		os.Exit(0)
	}

	// Call setsid
	_, err := syscall.Setsid()
	if err != nil {
		os.Exit(1)
	}

	// Fork again
	pid, _, _ = syscall.Syscall(syscall.SYS_FORK, 0, 0, 0)

	// Exit in parent again
	switch pid {
	case 0:
		// Child process, carry on
		break
	default:
		// Parent process, exit cleanly
		os.Exit(0)
	}

	// Clear umask
	syscall.Umask(0)

	// Change working directory
	err = syscall.Chdir("/")
	if err != nil {
		os.Exit(1)
	}

	// Duplicate /dev/null to stdin, stdout and stderr
	nullFile, err := os.OpenFile("/dev/null", os.O_RDWR, 0)
	if err != nil {
		os.Exit(1)
	}
	nullFd := nullFile.Fd()
	syscall.Dup2(int(nullFd), int(os.Stdin.Fd()))
	syscall.Dup2(int(nullFd), int(os.Stdout.Fd()))
	syscall.Dup2(int(nullFd), int(os.Stderr.Fd()))

}

// Handle HTTP errors here
func logHttpError(str string, rw http.ResponseWriter) {
	log.Printf("%s", str)
	var buffer strings.Builder
	buffer.WriteString("<!DOCTYPE html><html><head><title>")
	buffer.WriteString(str)
	buffer.WriteString("</title></head><body>")
	buffer.WriteString(str)
	buffer.WriteString("</body></html>")
	rw.WriteHeader(404)
	rw.Write([]byte(buffer.String()))
}

// This program listens for go get requests and returns the correct git URL for me
func main() {
	// Daemonise the process if flag is not present
	noDaemon := flag.Bool("no_daemon", false, "This flag, if true, stops the process from daemonising")
	flag.Parse()

	if !*noDaemon {
		daemonise()
	}

	// Define handler function for translating our Go import paths
	http.HandleFunc("/", func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Add("Content-Type", "text/html")

		// Input path should be of form "/module_path[/v1...]"
		path := req.URL.Path
		if len(path) == 0 {
			logHttpError("Path cannot be empty", rw)
			return
		}

		// Trim first '/'
		if path[0] != '/' {
			logHttpError("Path should begin with '/'", rw)
			return
		}
		path = path[1:]

		// Now split on remaining '/' characters if they exist - return slice will be at least size 1
		components := strings.Split(path, "/")
		if components[0] == "" {
			logHttpError("Module name cannot be empty", rw)
			return
		}

		// Create first part of GitHub path meta response
		var newPathBuffer strings.Builder
		newPathBuffer.WriteString("<!DOCTYPE html><html><head><meta name=\"go-import\" content=\"")
		newPathBuffer.WriteString(domainName)

		// Write path, and version number if specified - rest of components can be discarded as they refer to subpackages
		newPathBuffer.WriteString("/")
		newPathBuffer.WriteString(components[0])
		if len(components) > 1 && len(components[1]) > 0 {
			found, err := regexp.MatchString("v[0-9]+", components[1])
			if err == nil && found {
				// Add version number to path
				newPathBuffer.WriteString("/")
				newPathBuffer.WriteString(components[1])
			}
		}

		// Finish meta response
		newPathBuffer.WriteString(" git https://")
		newPathBuffer.WriteString(gitRoot)
		newPathBuffer.WriteString("/")
		newPathBuffer.WriteString(components[0])
		newPathBuffer.WriteString("\"></head><body></body></html>")

		// Write response
		_, err := rw.Write([]byte(newPathBuffer.String()))
		if err != nil {
			log.Printf("Couldn't write response to client")
			return
		}
	})
	log.Fatal(http.ListenAndServe("localhost:" + portNumber, nil))
}

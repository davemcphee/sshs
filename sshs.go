// sshs takes a single required argument, a hostname, and searches ~/.ssh/config for that hosts' info, and
// pretty prints it to stdout
// flags include --ssh_config.location for alternative config file to parse. See host_info --help for usage.
package main

import (
	"bufio"
	"fmt"
	"github.com/google/logger"
	"gopkg.in/alecthomas/kingpin.v2"
	"io/ioutil"
	"os"
	"strings"
)

// The struct that contains host data, if host was found
type hostInfo struct {
	hostname    string
	user        string
	ipAddr      string
	port        string
	keyfile     string
	bastionHost string // another hostInfo struct?
}

// Stringer interface for hostInfo struct
func (h hostInfo) String() string {
	return fmt.Sprintf("%s\nIP:\t\t%s\nUser:\t\t%s\nPort:\t\t%s\nJumphost:\t%s\n", h.hostname, h.ipAddr, h.user, h.port, h.bastionHost)
}

// parses an ssh config file, returns a hostInfo struct
func parseSSHConfig(fileLocation, hostName string) (hostInfo, error) {
	file, err := os.Open(fileLocation)

	if err != nil {
		return hostInfo{}, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	// Scan through config file line by line
	for scanner.Scan() {
		hostString := scanner.Text()

		// If the line looks like Host <our target host>, we may have a result
		if strings.Contains(hostString, "Host") && strings.Contains(hostString, hostName) {

			// init a hostStruct
			hostStruct := hostInfo{hostname: hostName}

			// scan next line
			scanner.Scan()

			// Inf loop until we find everything we want, or an empty line, indicating end of that host's section
			for {
				hostString = scanner.Text()

				switch {
				case strings.Contains(hostString, "User"):
					s := strings.Fields(hostString)
					hostStruct.user = s[1]
					scanner.Scan()
				case strings.Contains(hostString, "Hostname"):
					s := strings.Fields(hostString)
					hostStruct.ipAddr = s[1]
					scanner.Scan()
				case strings.Contains(hostString, "Port"):
					s := strings.Fields(hostString)
					hostStruct.port = s[1]
					scanner.Scan()
				case strings.Contains(hostString, "IdentityFile"):
					s := strings.Fields(hostString)
					hostStruct.keyfile = s[1]
					scanner.Scan()
				case strings.Contains(hostString, "ProxyCommand"):
					s := strings.Fields(hostString)
					hostStruct.bastionHost = strings.Join(s[1:], " ")
					scanner.Scan()
				case hostString == "":
					return hostStruct, nil
				// We found a line we don't care about - just scan to next line and loop
				default:
					scanner.Scan()
					hostString = scanner.Text()
				}
			}
		}
	}

	// Didn't find anything, and possibly got an error
	return hostInfo{}, scanner.Err()
}

func main() {
	// Get default config file location from $HOME
	userHome, _ := os.UserHomeDir()
	defaultSSHConfigPath := userHome + "/.ssh/config"

	var (
		sshConfigFile = kingpin.Flag(
			"conf_file",
			"path to ssh/config file to search",
		).Default(defaultSSHConfigPath).String()
		verbose = kingpin.Flag(
			"verbose",
			"lots o' logs",
		).Short('v').Bool()
		hostName = kingpin.Arg(
			"hostname",
			"host to search for (required)",
		).Required().String()
	)

	// parse cmd line args
	kingpin.Parse()

	// stdout logger
	stdoutLogger := logger.Init("Logger Example", *verbose, false, ioutil.Discard)
	stdoutLogger.Infof("Using config file %v", *sshConfigFile)

	// Parse the ssh config file
	host, err := parseSSHConfig(*sshConfigFile, *hostName)
	if err != nil {
		stdoutLogger.Errorf("Failed to parse config file: %v", err)
		os.Exit(1)
	}
	fmt.Println(host)
}

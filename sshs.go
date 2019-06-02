// sshs takes a single required argument, a hostname, and searches ~/.ssh/config for that hosts' info, and
// pretty prints it to stdout
// flags include --conf_file for alternative config file to parse. See sshs --help for usage.
package main

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/google/logger"
	"gopkg.in/alecthomas/kingpin.v2"
	"io/ioutil"
	"os"
	"strings"
)

// Exported interface for hostInfo struct
type HostInfo interface {
	String() string
}

// The struct that contains host data, if host was found. It's internal, and exported via it's interface HostInfo
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

// hostInfo factory method - takes []string, returns new HostInfo (the interface, not the private struct)
func NewHostInfo(chunk []string) (HostInfo, error) {
	li := hostInfo{}
	findCounter := 0

	for _, l := range chunk {
		switch {
		case strings.Contains(l, "Host "):
			s := strings.Fields(l)
			li.hostname = s[1]
			findCounter++
		case strings.Contains(l, "User"):
			s := strings.Fields(l)
			li.user = s[1]
			findCounter++
		case strings.Contains(l, "Hostname"):
			s := strings.Fields(l)
			li.ipAddr = s[1]
			findCounter++
		case strings.Contains(l, "Port"):
			s := strings.Fields(l)
			li.port = s[1]
			findCounter++
		case strings.Contains(l, "IdentityFile"):
			s := strings.Fields(l)
			li.keyfile = s[1]
			findCounter++
		case strings.Contains(l, "ProxyCommand"):
			s := strings.Fields(l)
			li.bastionHost = strings.Join(s[1:], " ")
			findCounter++
		}
	}
	if findCounter < 5 {
		return li, errors.New("failed to build hostInfo struct from input")
	}
	return li, nil
}

// Parse a file in chunks and send data to a channel
func parseChunks(fileLocation string, c chan []string, stopChan chan bool) {
	file, err := os.Open(fileLocation)

	if err != nil {
		close(c)
		return
	}
	defer file.Close()

	var block []string
	var blockCounter uint16

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		select {
		// this chan is empty, so the read blocks, so select skips it and goes to default - until it's closed, then it's
		// able to read an error, and quit
		case <-stopChan:
			// fmt.Printf("parseChunks() got stop signal, exiting after %d blocks\n", blockCounter)
			return
		default:
			l := scanner.Text()

			// we got a line, append to block slice
			if len(l) != 0 {
				block = append(block, l)
				continue
			}

			// we got a "" and it's the end of a block; send block to chan and clear block
			if len(l) == 0 && len(block) != 0 {
				blockCounter++
				c <- block
				block = []string{}
				continue
			}

			// empty string and block is empty = double empty?
			if len(l) == 0 {
				block = []string{}
				continue
			}
		}
	}
	// We're done parsing file, close channel
	// fmt.Printf("parseChunks() exiting at EOF; no stop signal received.")
	close(c)
}

// receive chunks, determine if it's the one we want, sends result to resChan
func chunkReader(hostName string, chunkChan chan []string, stopChan chan bool, resChan chan HostInfo) {
	for chunk := range chunkChan {
		// fmt.Println("Reading a chunk...")
		if strings.Contains(chunk[0], hostName) {
			// fmt.Println("Got it! Sending stop signal to stopChan.")
			close(stopChan)
			newHostInfo, _ := NewHostInfo(chunk)
			resChan <- newHostInfo
			close(resChan)
		}
	}
	fmt.Println("Didn't find hostName")
	close(chunkChan)
	close(resChan)
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

	// Create a channel to read result on
	chunkChan := make(chan []string)

	// Channel to stop threads
	stopChan := make(chan bool)

	// Channel to read result
	resultChan := make(chan HostInfo)

	// parseChunks sends chunks of file to channel, until it's told to stop on stopChan
	go parseChunks(*sshConfigFile, chunkChan, stopChan)

	// reads from chunkChan until it finds our hostname, then sends stop to stopChan
	go chunkReader(*hostName, chunkChan, stopChan, resultChan)

	res := <-resultChan
	fmt.Println(res)
}

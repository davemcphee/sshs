package main

import (
	"log"
	"os"
)

import "io/ioutil"
import "testing"

const selma0004 string = `
Host selma0004
	User root
	Hostname 10.1.0.156
	Port 22
	Protocol 2
	IdentityFile /home/alex/.ssh/gitolite
	ProxyCommand none

Host dirtycid00024
    User root
    Hostname 10.1.2.7
    Port 22
    Protocol 2
    IdentityFile /home/alex/.ssh/gitolite
    ProxyCommand ssh -W %h:%p -q root@52.144.40.132
`

func TestSSHConfigFileParsing(t *testing.T) {
	content := []byte(selma0004)

	tmpfile, err := ioutil.TempFile("", "host_info.*.test")
	if err != nil {
		log.Fatal(err)
	}

	defer os.Remove(tmpfile.Name()) // clean up

	if _, err := tmpfile.Write(content); err != nil {
		log.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		log.Fatal(err)
	}

	// Test selma0004
	host, err := parseSSHConfig(tmpfile.Name(), "selma0004")
	if err != nil {
		t.Errorf("parseSSHConfig(selma0004) failed: %v", err)
	}

	if host.ipAddr != "10.1.0.156" {
		t.Errorf("Got wrong data from parseSSHConfig(selma0004): ipAddr is %v", host.ipAddr)
	}

	if host.bastionHost != "none" {
		t.Errorf("Got wrong bastionhost from parseSSHConfig(selma0004): %v", host.bastionHost)
	}

	// Test dirtycid00024
	host2, err2 := parseSSHConfig(tmpfile.Name(), "dirtycid00024")
	if err2 != nil {
		t.Errorf("parseSSHConfig(dirtycid00024) failed: %v", err2)
	}

	if host2.ipAddr != "10.1.2.7" {
		t.Errorf("Got wrong data from parseSSHConfig(dirtycid00024): ipAddr is %v", host2.ipAddr)
	}

	if host2.bastionHost != "ssh -W %h:%p -q root@52.144.40.132" {
		t.Errorf("Got wrong bastionhost from parseSSHConfig(dirtycid00024): %v", host2.bastionHost)
	}
}

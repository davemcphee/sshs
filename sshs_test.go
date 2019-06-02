package main

import (
	"os"
)

import "io/ioutil"
import "testing"

const selma0004 string = `
Host server004
	User root
	Hostname 10.1.2.3
	Port 22
	Protocol 2
	IdentityFile /home/user/.ssh/id_rsa
	ProxyCommand none

Host server00024
    User root
    Hostname 10.1.2.3
    Port 22
    Protocol 2
    IdentityFile home/user/.ssh/id_rsa
    ProxyCommand ssh -W %h:%p -q root@1.2.3.4

`

func Test_parseChunks(t *testing.T) {
	t.Run("parseChunks reading from tmp file", func(t *testing.T) {
		content := []byte(selma0004)

		tmpfile, err := ioutil.TempFile("", "host_info.*.test")
		if err != nil {
			t.Error(err)
		}

		defer os.Remove(tmpfile.Name()) // clean up

		if _, err := tmpfile.Write(content); err != nil {
			t.Error(err)
		}
		if err := tmpfile.Close(); err != nil {
			t.Error(err)
		}

		// set up some channels and test parseChunks

		chunkChan := make(chan []string)
		stopChan := make(chan bool)

		go parseChunks(tmpfile.Name(), chunkChan, stopChan)

		chunkCounter := 0

		for range chunkChan {
			chunkCounter++
		}

		if chunkCounter != 2 {
			t.Error("Failed to get 2 chunks from parseChunks()")
		}
	})
}

func Test_NewHostInfo(t *testing.T) {
	t.Run("Make HostInfo struct with bad data", func(t *testing.T) {
		if _, err := NewHostInfo([]string{"", "", "", "", ""}); (err == nil) != false {
			t.Error("NewHostInfo didn't return error when it should have")
		}
	})
	t.Run("Make HostInfo struct with good data", func(t *testing.T) {
		dats := []string{"Host selma0004", "User root", "Hostname 10.1.0.156",
			"Port 22", "Protocol 2", "IdentityFile /home/alex/.ssh/gitolite",
			"ProxyCommand none"}
		if _, err := NewHostInfo(dats); (err != nil) != false {
			t.Errorf("NewHostInfo returned error: %v", err)
		}
	})
}

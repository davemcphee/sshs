[![Build Status](https://travis-ci.org/davemcphee/sshs.svg?branch=master)](https://travis-ci.org/davemcphee/sshs)

~~~~
$ sshs servername00024
servername00024
IP:		10.1.2.3
User:		root
Port:		22
Jumphost:	root@123.4.5.6
~~~~

~~~~
usage: sshs [<flags>] <hostname>

Flags:
  --help  Show context-sensitive help (also try --help-long and --help-man).
  --conf_file="/home/alex/.ssh/config"  
          path to ssh/config file to search

Args:
  <hostname>  host to search for (required)
~~~~

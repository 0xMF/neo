package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var course = ""
var ground = ""
var season = ""
var secret = ""
var server = []string{}

var folder = ""
var adminF = ""
var entryF = ""

var message = ""
var modules = ""
var replyTo = ""

// --- x ---

var askDir = ""
var logDir = ""
var errDir = ""

var shDone = folder + "/shDone "
var shInit = folder + "/shInit " + adminF
var shLkup = folder + "/shLkup "
var shMail = folder + "/shMail " + adminF
var shWins = folder + "/shWins "

var logRWrite os.FileMode = 0244
var logUpdate os.FileMode = 0644

func verify() {

	srv := "/bin/hostname | /usr/bin/awk -F. '{print $(NF-1),$NF}' | /bin/sed 's| |.|g'"
	c := " | /usr/bin/cut -d'/' -f1-3"

	// verify server
	cmd := exec.Command("/bin/bash", "--norc", "--noprofile", "-c", srv)
	stdin, err := cmd.StdinPipe()
	EOK(errDir, err, fmt.Sprintf("Cannot create pipe to check server"))
	defer stdin.Close()

	out, err := cmd.CombinedOutput()
	EOK(errDir, err, fmt.Sprintf("Cannot get output to verify server authenticity"))
	s := strings.TrimSuffix(string(out), "\n")
	auth := false
	for _,v := range server {
		if s == v {
			auth = true
			break
		}
	}
	if ! auth {
		EOK(errDir, errors.New("Unable to confirm server: "+usrname+" "+s))
	}

	// verify path
	ex, err := os.Executable()
	p, err := filepath.EvalSymlinks(ex)
	c = "/bin/echo " + p + c
	cmd = exec.Command("/bin/bash", "--norc", "--noprofile", "-c", c)
	stdin, err = cmd.StdinPipe()
	EOK(errDir, err, fmt.Sprintf("Cannot create pipe to check path %s", usrname))
	defer stdin.Close()

	out, err = cmd.CombinedOutput()
	EOK(errDir, err, fmt.Sprintf("Cannot get output to verify path authenticity"))
	s = strings.TrimSuffix(string(out), "\n")
	if s != ground {
		EOK(errDir, errors.New("Unable to verify path: "+usrname+" "+s))
	}
}

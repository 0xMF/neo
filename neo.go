package main

import (
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
)

var mission = "default"

func main() {
	log.Printf(string(os.Args[0]))
	exe, err := filepath.EvalSymlinks(os.Args[0])
	mission = path.Dir(exe) + "/" + mission
	log.Printf("Start: %v", err)

	mCmd := exec.Command("/bin/bash", "-c", mission)
	mCmd.Stdin = os.Stdin
	mCmd.Stdout = os.Stdout
	mCmd.Stderr = os.Stderr
	mCmd.Start()
	err = mCmd.Wait()
	log.Printf("Stop : %v", err)
}

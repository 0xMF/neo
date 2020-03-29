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
	mission = path.Dir(exe) + "/.0xMF/" + mission
	log.Printf("Begin: %v", err)

	mCmd := exec.Command("/bin/bash", "-c", mission)
	mCmd.Stdin = os.Stdin
	mCmd.Stdout = os.Stdout
	mCmd.Stderr = os.Stderr
	mCmd.Start()
	err = mCmd.Wait()
	log.Printf("End  : %v", err)
}

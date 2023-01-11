package main

import (
	"encoding/json"
	"io"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
)

var config = ".neo.json"
var mission = "default"

type Location struct {
	Name string `json:"name"`
	Dest string `json:"dest"`
	Path string `json:"path"`
}

func Resolve(h, p, w string) error {
	mission = h + "/" + p + "/" + w + "/s/" + mission

	mCmd := exec.Command("/bin/bash", "-c", mission)
	mCmd.Stdin = os.Stdin
	mCmd.Stdout = os.Stdout
	mCmd.Stderr = os.Stderr
	mCmd.Start()
	return mCmd.Wait()
}

// parse (JSON) config and if name matching this exectuable is found, dispatch to resolver
func Dispatch(d, e string) error {
	var data []Location

	config = d + "/" + config
	jsonF, _ := os.Open(config)
	defer jsonF.Close()
	plan, _ := io.ReadAll(jsonF)
	if json.Unmarshal(plan, &data) != nil {
		log.Fatal("could not parse config")
	}

	for _, l := range data {
		if e == l.Name {
			return Resolve(d, l.Path, l.Dest)
		}
	}
	log.Fatal("not ready yet.")
	return nil
}

func main() {
	log.Printf(string(os.Args[0]))

	abs, err := filepath.EvalSymlinks(os.Args[0])
	dir := path.Dir(abs)
	exe := path.Base(abs)

	log.Printf("Begin: %v", err)
	err = Dispatch(dir, exe)
	log.Printf("End  : %v", err)
}

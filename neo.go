package main

import (
	"log"
	"os"
	"os/exec"
	"sync"
)
var	neo = make(chan int)
var mission = "default"
var pwd string

func main() {
	var err error
	pwd, err = os.Getwd()
	log.Printf("Begin mission: %v", err)

	var wg sync.WaitGroup
	wg.Add(1)
	go workerItem(1,&wg)
	<-neo
	wg.Wait()
}

func workerItem(id int, wg *sync.WaitGroup) {

	defer wg.Done()

	mission = pwd + "/" + mission
	log.Printf(mission)

	cmd := exec.Command("/bin/bash", "-c", mission)
	err := cmd.Run() // cmd.Run()
	log.Printf("Finished checking item: %v", err)
	neo <- 1
}

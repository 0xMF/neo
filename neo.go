package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

var mission = "/usr/bin/" + item

func main() {
}

func workerItem(id int, wg *sync.WaitGroup) {

	defer wg.Done()

	itemCheckPost = itemCheckPost + " " + localF
	//fmt.Println(itemCheckPost)

	cmdline := exec.Command("/bin/bash", "-c", itemCheckPost)
	err := cmdline.Run()
	log.Printf("Finished checking item: %v", err)
	neo <- 1
}

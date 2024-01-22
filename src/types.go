package main

import (
	"encoding/csv"
	"sync"
	"time"

	"github.com/awesome-gocui/gocui"
)

// fields must be public (uppercase) for unmarshal to correctly populate the data.
type Test struct {
	Topic     string `yaml:"topic"`
	Author    string
	Update    string
	Questions []QSet `yaml:"questions"`
}

type Player struct {
	Name string
	Team string
	Lead string
	Done int
}

type Complete struct {
	Modules map[int]bool
}

type QSet struct {
	Ask string   `yaml:"ask"`
	Ans []string `yaml:"ans"`
}

type Terminal struct {
	*gocui.Gui
	views         map[string]handle
	height, width int
}

type handle struct {
	*gocui.View
	text string
}

var counter = 0
var csvDone = new(csv.Writer)
var csvFile = new(csv.Writer)
var csvStats = new(csv.Writer)
var done Complete
var doneF string
var mdStart, mdEnd time.Time
var pDetails string
var player Player
var term Terminal
var test = Test{}
var topicNo = 0
var usrname string
var wg sync.WaitGroup
var ymlFile = "0"

var version = "neo version 1.5.3-beta by Mark Fernandes on 2024-Jan-22."

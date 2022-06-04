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
	Name  string
	Team  string
	Lead  string
	Score int
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

var mdStart, mdEnd time.Time
var counter = 0
var csvFile = new(csv.Writer)
var csvDone = new(csv.Writer)
var usrname string
var topicNo = 0
var ymlFile = "0"
var done Complete
var doneF string
var term Terminal
var test = Test{}
var csvStats = new(csv.Writer)
var wg sync.WaitGroup
var player Player
var pDetails string

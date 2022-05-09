// +build neo

package main

import (
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/jroimartin/gocui"
	"gopkg.in/yaml.v2"
)

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

// fields must be public (uppercase) for unmarshal to correctly populate the data.
type Test struct {
	Topic     string `yaml:"topic"`
	Author    string
	Update    string
	Questions []QSet `yaml:"questions"`
}

type Player struct {
	Name		string
	Team		string
	Lead		string
	Score		int
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
	call func(*gocui.View) error
	text string
}

func initDone() {

	var errBytes bytes.Buffer
	cmd := exec.Command("/bin/bash", "-c", shInit)
	cmd.Stdin		= os.Stdin
	cmd.Stdout	= os.Stdout
	cmd.Stderr  = &errBytes
	err := cmd.Start()
	EOK(errDir, err, "couldn't run init script", errBytes.String())
	err = cmd.Wait()
	EOK(errDir, err, "couldn't wait", errBytes.String())
	var result	=	strings.Split(errBytes.String(), ",")
	player.Team = result[0]
	player.Lead = result[1]
	player.Score = 0
	EOK(errDir, err, "did not finish script", errBytes.String())
}

func sendMail() {

	var errBytes bytes.Buffer
	doneF = logDir + "/" + player.Name + ".done"
	var shMail = shMail + " " + player.Team + " " + doneF
	cmd := exec.Command("/bin/bash", "-c", shMail)
	cmd.Stdin  = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = &errBytes
	err := cmd.Start()
	EOK(errDir, err, "mail not sent", errBytes.String())
	err = cmd.Wait()
	EOK(errDir, err, "couldn't send email", errBytes.String())
}


func main() {

	u, err := user.Current(); OK(err)
	usrname = u.Username
	player.Name = u.Username

	l, err := filepath.EvalSymlinks(logDir)
	EOK(errDir,err, fmt.Sprintf("Cannot open logDir"))

	l += "/" + usrname
	if _, err := os.Stat(l); errors.Is(err, os.ErrNotExist) {
		// initial setup
		fmt.Println("First time here eh?")
		initDone()
	}

	// detailed log
	file, err := os.OpenFile(l, os.O_APPEND|os.O_CREATE|os.O_WRONLY, logRWrite)
	EOK(errDir, err, fmt.Sprintf("Cannot access log for user: %s", usrname))
	defer file.Close()

	err = os.Chmod(l, logRWrite)
	EOK(errDir, err, fmt.Sprintf("Log permission error for %s", usrname))

	csvFile = csv.NewWriter(file)
	message := []string{string(time.Now().Format(time.RFC822)), player.Team, usrname, "---"}
	csvFile.Write(message)
	csvFile.Flush()
	EOK(errDir, csvFile.Error(), fmt.Sprintf("Cannot create csv files"))
	log.SetOutput(os.Stdout)

	// done log
	done.Modules = make(map[int]bool)
	l += ".done"
	file, err = os.OpenFile(l, os.O_CREATE|os.O_APPEND|os.O_RDWR, logUpdate)
	err = os.Chmod(l, logUpdate)
	EOK(errDir,err, fmt.Sprintf("Cannot update stats for user: %s", usrname))
	defer file.Close()

	csvStats = csv.NewWriter(file)

	readTest(ymlFile)

	term.Gui, err = gocui.NewGui(gocui.OutputNormal)
	OK(err)
	defer term.Gui.Close()

	term.Gui.Cursor = true
	term.width, term.height = term.Gui.Size()

	term.views = make(map[string]handle)
	term.views["header"] = handle{call: header, text: "Topic"}
	term.views["writer"] = handle{call: writer, text: "Question"}
	term.views["reader"] = handle{call: reader, text: ""}
	term.views["status"] = handle{call: status, text: "Ready."}

	term.Gui.SetManagerFunc(layout)
	EOK(errDir,keybindings(term.Gui), fmt.Sprintf("Cannot set keybindings for %s", usrname))

	if err := term.Gui.MainLoop(); err != nil && err != gocui.ErrQuit && err != file.Sync() {
		log.Panicln(err)
	}
}

func isDone(b bool) string {
	if b {
		return "Y"
	}
	return "N"
}

func check(g *gocui.Gui, v *gocui.View) error {
	if counter+1 >= len(test.Questions) || test.Topic == "Menu" {
		if counter+1 >= len(test.Questions) && test.Topic != "Menu" {
			message := []string{string(time.Now().Format(time.RFC822)), player.Name, player.Team, ymlFile, test.Topic, "ENDED"}
			csvFile.Write(message)
			csvFile.Flush()

			if !done.Modules[topicNo] {
				done.Modules[topicNo] = true
				for i, d := range done.Modules {
					message := []string{string(time.Now().Format(time.RFC822)),
					player.Name, player.Team, player.Lead, strconv.Itoa(player.Score), strconv.Itoa(i), isDone(d)}
					csvStats.Write(message)
					csvStats.Flush()
				}
				OK(csvStats.Error())
				log.SetOutput(os.Stdout)
				sendMail()
			}
		}
		return refreshMenu(g, v)
	}

	yes := "CORRECT! Press Ctrl+D to exit or Enter to continue."
	in := response(v)
	if isAnswer(in) { // correct answers show both: answer and congrats message
		if counter == 0 || term.views["status"].text == yes {
			ok(refresh("reader", g, ""))
			ask(g, v)
		} else {
			if term.views["status"].text != yes {
				ok(refresh("status", g, yes))
			}
		}

	} else { // toggle incorrect statuses between 'Ready' and 'Try again'
		ok(refresh("reader", g, ""))
		if term.views["status"].text == "Ready." {
			ok(refresh("status", g, "Incorrect. Try again."))
		} else { // empty input (hitting ENTER) clears status message
			if in == "" {
				ok(refresh("status", g, "Ready."))
			}
		}
	}
	return nil
}

func isAnswer(in string) bool {
	if len(test.Questions[counter].Ans) == 1 {
		return in == test.Questions[counter].Ans[0]
	}
	for _, answer := range test.Questions[counter].Ans {
		if in == answer {
			return true
		}
	}
	return false
}

func ask(g *gocui.Gui, v *gocui.View) {
	counter = counter + 1
	ok(refresh("header", g, "("+string(counter+1)+"/"+string(len(test.Questions))+") "+test.Topic))
	ok(refresh("writer", g, test.Questions[counter].Ask))
	ok(refresh("status", g, "Ready."))
}

func readTest(y string) {
	q, err := filepath.EvalSymlinks(askDir)
	OK(err, "ask directory not found")
	data, err := ioutil.ReadFile(q + "/" + y + ".yaml")
	if err != nil {
		data, err = ioutil.ReadFile(q + "/0" + y + ".yaml")
	}
	OK(err)

	ok(yaml.Unmarshal([]byte(data), &test))
}

func refreshMenu(g *gocui.Gui, v *gocui.View) error {
	refreshTerm(g)
	selectMenu(g, v)
	return nil
}

func refreshTerm(g *gocui.Gui) {
	counter = 0
	topicNo = 0
	ymlFile = "0"
	readTest(ymlFile)
	ok(refresh("header", g, test.Topic))
	ok(refresh("writer", g, test.Questions[counter].Ask))
	ok(refresh("reader", g, ""))
	ok(refresh("status", g, "Ready."))
}

func response(v *gocui.View) string {
	_, cy := v.Cursor()
	in, err := v.Line(cy)
	if err != nil {
		in = ""
	}
	return in
}

func selectMenu(g *gocui.Gui, v *gocui.View) {
	in := response(v)
	ok(refresh("reader", g, ""))

	if isAnswer(in) { // user selected existing menu item, so read that item's test paper
		ymlFile = in
		topicNo, _ = strconv.Atoi(ymlFile)
		readTest(ymlFile)

		message := []string{string(time.Now().Format(time.RFC822)), player.Name, player.Team, ymlFile, test.Topic, "BEGAN"}
		csvFile.Write(message)
		csvFile.Flush()

		ok(refresh("header", g, "("+string(counter+1)+"/"+string(len(test.Questions))+") "+test.Topic))
		ok(refresh("writer", g, test.Questions[counter].Ask))

	} else { // user entered invalid menu selection, ignore
		ok(refresh("reader", g, ""))
		if in == "" {
			ok(refresh("status", g, "Ready."))
		} else {
			ok(refresh("status", g, "Incorrect. Try again."))
		}
	}
	in = ""
}

//go:build neo
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
	"time"

	"github.com/awesome-gocui/gocui"
	"gopkg.in/yaml.v2"
)

func main() {

	mdStart = time.Now()
	mdEnd = time.Now()
	u, err := user.Current()
	OK(err)
	usrname = u.Username
	player.Name = u.Username

	l, err := filepath.EvalSymlinks(logDir)
	EOK(errDir, err, fmt.Sprintf("Cannot open logDir"))

	l += "/" + usrname

	if _, err := os.Stat(l); errors.Is(err, os.ErrNotExist) { // new player setup
		player.Team = "nil"
		player.Lead = "nil"
	}

	// details log
	file, err := os.OpenFile(l, os.O_APPEND|os.O_CREATE|os.O_WRONLY, logRWrite)
	EOK(errDir, err, fmt.Sprintf("Cannot access log for user: %s", usrname))
	defer file.Close()
	err = os.Chmod(l, logRWrite)
	EOK(errDir, err, fmt.Sprintf("Log permission error for %s", usrname))
	csvFile = csv.NewWriter(file)

	fi, _ := file.Stat()
	if fi.Size() <= 10 {
		player.Team = "nil"
		player.Lead = "nil"
	}

	// done log
	done.Modules = make(map[int]bool)
	l += ".done"
	file, err = os.OpenFile(l, os.O_CREATE|os.O_APPEND|os.O_RDWR, logUpdate)
	err = os.Chmod(l, logUpdate)
	EOK(errDir, err, fmt.Sprintf("Cannot update stats for user: %s", usrname))
	defer file.Close()
	csvStats = csv.NewWriter(file)

	fi, _ = file.Stat()
	if fi.Size() <= 10 {
		player.Team = "nil"
		player.Lead = "nil"
	}

	updatePlayer()
	if player.Team == "nil" || player.Lead == "nil" {
		ymlFile = levelX
	} else {
		ymlFile = "0"
	}

	message := []string{string(time.Now().Format(time.RFC822)), player.Team, usrname, "---"}
	csvFile.Write(message)
	csvFile.Flush()
	EOK(errDir, csvFile.Error(), fmt.Sprintf("Cannot create csv files"))
	log.SetOutput(os.Stdout)

	readTest(ymlFile)
	addPlayerDetails()

	term.Gui, err = gocui.NewGui(gocui.OutputNormal, true)
	OK(err)
	defer term.Gui.Close()

	term.Gui.Cursor = true
	term.width, term.height = term.Gui.Size()

	term.views = make(map[string]handle)
	term.views["header"] = handle{text: version}
	if data, err := os.ReadFile(entryF); err == nil {
		term.views["writer"] = handle{text: string(data)}
	} else {
		term.views["writer"] = handle{text: "\n\t\tAre you ready to play?\n\t\tUse CTRL+D to exit and ENTER to begin"}
	}
	term.views["reader"] = handle{text: ""}
	term.views["status"] = handle{text: "Press ENTER to begin."}

	term.Gui.SetManagerFunc(layout)
	EOK(errDir, keybindings(term.Gui), fmt.Sprintf("Cannot set keybindings for %s", usrname))

	if err := term.Gui.MainLoop(); err != nil && err != gocui.ErrQuit && err != file.Sync() {
		log.Panicln(err)
	}
}

func addPlayerDetails() {
	test.Questions[0].Ask += pDetails
}

func lookUp(t string) string {

	var errBytes bytes.Buffer
	lookup := shLkup + " " + t + " " + adminF
	cmd := exec.Command("/usr/bin/bash", "--noprofile", "-c", lookup)
	stdin, err := cmd.StdinPipe()
	EOK(errDir, err, "couldn't create input pipe", errBytes.String())
	defer stdin.Close()
	out, err := cmd.CombinedOutput()
	EOK(errDir, err, "couldn't start stdin", errBytes.String())
	return (strings.TrimSuffix(string(out), "\n"))
}

func completed() int {

	var errBytes bytes.Buffer
	finished := shDone + " " + doneF
	cmd := exec.Command("/usr/bin/bash", "--noprofile", "-c", finished)
	stdin, err := cmd.StdinPipe()
	EOK(errDir, err, "couldn't create input pipe", errBytes.String())
	defer stdin.Close()
	out, err := cmd.CombinedOutput()
	EOK(errDir, err, "couldn't start stdin", errBytes.String())
	n, err := strconv.Atoi(strings.TrimSuffix(string(out), "\n"))
	EOK(errDir, err, "couldn't convert stdin", errBytes.String())
	return n
}

func initDone2(s int) {

	var errBytes bytes.Buffer
	cmd := exec.Command("/usr/bin/bash", "--noprofile", "-c", shDone)
	stdin, err := cmd.StdinPipe()
	EOK(errDir, err, "couldn't create input pipe", errBytes.String())
	defer stdin.Close()
	out, err := cmd.CombinedOutput()
	EOK(errDir, err, "couldn't start stdin", errBytes.String())
	var result = strings.Split(string(out), ",")
	log.Fatal(result)
	player.Team = result[0]
	//player.Lead = result[1]
	player.Lead = lookUp(player.Team)
	player.Score = s
	EOK(errDir, err, "did not finish script", errBytes.String())
}

func initDone(s int) {

	var errBytes bytes.Buffer
	cmd := exec.Command("/usr/bin/bash", "--noprofile", "-c", shInit)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = &errBytes
	err := cmd.Start()
	EOK(errDir, err, "couldn't run init script", errBytes.String())
	err = cmd.Wait()
	EOK(errDir, err, "couldn't wait", errBytes.String())

	var result = strings.Split(errBytes.String(), ",")
	if len(result) != 2 {
		EOK(errDir, errors.New("can't begin"), "error at init", errBytes.String())
	}
	player.Team = result[0]
	player.Lead = result[1]
	player.Score = s
	EOK(errDir, err, "did not finish script", errBytes.String())
}

func sendMail(s string) {

	var errBytes bytes.Buffer
	doneF = logDir + "/" + player.Name + ".done"
	mail := shMail + " " + player.Name + " " + teamPre + " " + player.Team + " " + s + " " + subPost + " " + doneF
	cmd := exec.Command("/usr/bin/bash", "--noprofile", "-c", mail)
	//log.Println(cmd.String()[42:])
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = &errBytes
	err := cmd.Start()
	EOK(errDir, err, "mail not sent", errBytes.String())
	err = cmd.Wait()
	EOK(errDir, err, "couldn't send email", errBytes.String())
}

func updatePlayer() {

	if player.Team == "nil" {
		initDone(0)
		zero := strconv.Itoa(player.Score)
		// date,user,team,leader,score,level,time
		message := []string{string(mdEnd.Format(time.RFC822)),
			player.Name, player.Team, player.Lead, zero, "", zero + "s"}
		csvStats.Write(message)
		csvStats.Flush()
	}

	var outBytes bytes.Buffer
	doneF = logDir + "/" + player.Name + ".done"
	cli := "/usr/bin/tail -1 " + doneF
	cmd := exec.Command("/usr/bin/bash", "--noprofile", "-c", cli)
	cmd.Stdin = os.Stdin
	cmd.Stdout = &outBytes
	cmd.Stderr = &outBytes
	err := cmd.Start()
	EOK(errDir, err, "couldn't get  last line", outBytes.String())
	err = cmd.Wait()
	EOK(errDir, err, "couldn't finish last line", outBytes.String())
	var result = strings.Split(outBytes.String(), ",")
	if len(outBytes.String()) > 0 {
		player.Team = result[2]
		//player.Lead = result[3]
		player.Lead = lookUp(player.Team)
		player.Score, _ = strconv.Atoi(result[4])
		pDetails = "\n\n\t=========================================================================="
		pDetails += "\n\n\tYou are user " + player.Name + " in section " + player.Team + " with "
		if player.Score == 1 {
			pDetails += player.Lead + ".\n\tYou completed " + strconv.Itoa(player.Score) + " module: "
		} else {
			if player.Score == 0 {
				pDetails += player.Lead + ".\n\tYou completed " + strconv.Itoa(player.Score) + " modules."
			} else {
				pDetails += player.Lead + ".\n\tYou completed " + strconv.Itoa(player.Score) + " modules: "
			}
		}
	}

	if player.Score > 0 {
		outBytes.Reset()
		wins := shWins + " " + doneF
		cmd = exec.Command("/usr/bin/bash", "--noprofile", "-c", wins)
		cmd.Stdin = os.Stdin
		cmd.Stdout = &outBytes
		cmd.Stderr = &outBytes
		err = cmd.Start()
		EOK(errDir, err, "couldn't start wins", outBytes.String())
		err = cmd.Wait()
		EOK(errDir, err, "couldn't finish wins", outBytes.String())
		result = strings.Split(outBytes.String(), ",")
		for _, si := range result {
			i, _ := strconv.Atoi(si)
			done.Modules[i] = true
			pDetails += si + " "
		}
	}
	pDetails += "\n\n\t=========================================================================="
}

func isDone(b bool) string {
	if b {
		return "Y"
	}
	return "N"
}

func timeTaken(end time.Time) string {
	elapsed, _ := time.ParseDuration(end.Sub(mdStart).String())
	return fmt.Sprintf("%s", elapsed.Round(time.Second).String())
}

func check(g *gocui.Gui, v *gocui.View) error {
	if test.Topic == "Team Selection" {
		done.Modules[topicNo] = true
		return showTeam(v)
	}

	if test.Topic == "Menu" {
		return showMenu(v)
	}

	if counter+1 >= len(test.Questions) && test.Topic != "Menu" {
		module := strconv.Itoa(topicNo)
		message := []string{string(time.Now().Format(time.RFC822)), player.Name, player.Team, ymlFile, test.Topic, "ENDED"}
		csvFile.Write(message)
		csvFile.Flush()

		if !done.Modules[topicNo] {
			done.Modules[topicNo] = true
			player.Score += 1
		}
		mdEnd = time.Now() // stop timer and write: date,user,team,leader,score,level,time
		message = []string{string(mdEnd.Format(time.RFC822)),
			player.Name, player.Team, player.Lead, strconv.Itoa(player.Score), module, timeTaken(mdEnd)}
		csvStats.Write(message)
		csvStats.Flush()
		updatePlayer()
		OK(csvStats.Error())
		log.SetOutput(os.Stdout)
		module += "/" + levels
		sendMail(module)
		mdStart = time.Now()
		resetTerm("0")
	} else {
		return checkResponse(g, v)
	}

	return nil
}

func checkResponse(g *gocui.Gui, v *gocui.View) error {

	yes := "CORRECT! Press Enter to continue or Ctrl+D to exit."
	in := response(v)
	if isAnswer(in) { // correct answers show both: answer and congrats message
		if counter == 0 || term.views["status"].text == yes {
			refresh("reader", "")
			ask(g, v)
		} else {
			if term.views["status"].text != yes {
				refresh("status", yes)
			}
		}

	} else { // toggle incorrect statuses between 'Ready' and 'Try again'
		refresh("reader", "")
		if term.views["status"].text == "Ready." {
			refresh("status", "Incorrect. Try again.")
		} else { // empty input (hitting ENTER) clears status message
			if in == "" {
				refresh("status", "Ready.")
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
	refresh("header", "("+strconv.Itoa(counter+1)+"/"+strconv.Itoa(len(test.Questions))+") "+test.Topic)
	refresh("writer", test.Questions[counter].Ask)
	refresh("status", "Ready.")
}

func readTest(y string) {
	q, err := filepath.EvalSymlinks(askDir)
	EOK(errDir, err, "ask "+y+" not found")
	data, err := ioutil.ReadFile(q + "/" + y + ".yaml")
	if err != nil {
		data, err = ioutil.ReadFile(q + "/0" + y + ".yaml")
	}
	EOK(errDir, err, "ask 0"+y+" not found")

	EOK(errDir, yaml.Unmarshal([]byte(data), &test), y+" can't unmarshal")
}

func resetFrame() {
	for n := range term.views {
		v := term.views[n].View
		v.Frame = true
		if v.FrameColor == gocui.ColorCyan {
			v.FrameColor = gocui.ColorGreen
		} else {
			v.FrameColor = gocui.ColorCyan
		}
	}
}

func resetTerm(y string) {
	counter = 0
	topicNo = 0
	ymlFile = y

	if player.Name == "nil" || player.Team == "nil" {
		ymlFile = levelX
	}
	readTest(y)
	addPlayerDetails()
	refresh("header", "("+strconv.Itoa(counter+1)+"/"+strconv.Itoa(len(test.Questions))+") "+test.Topic)
	refresh("writer", test.Questions[counter].Ask)
	refresh("reader", "")
	refresh("status", "Ready.")
	resetFrame()
}

func response(v *gocui.View) string {
	_, cy := v.Cursor()
	in, err := v.Line(cy)
	if err != nil {
		in = ""
	}
	return in
}

func showMenu(v *gocui.View) error {
	in := response(v)

	if player.Name == "nil" || player.Team == "nil" || in == "S" {
		ymlFile = levelX
		resetTerm(ymlFile)
		return showTeam(v)
	}

	ymlFile = "0"
	resetTerm(ymlFile)

	// user selected existing menu item, so read that item's module
	if isAnswer(in) {
		ymlFile = in
		topicNo, _ = strconv.Atoi(ymlFile)
		readTest(ymlFile)

		mdStart = time.Now() // start timer
		message := []string{string(mdStart.Format(time.RFC822)), player.Name, player.Team, ymlFile, test.Topic, "BEGAN"}
		csvFile.Write(message)
		csvFile.Flush()

		refresh("header", "("+strconv.Itoa(counter+1)+"/"+strconv.Itoa(len(test.Questions))+") "+test.Topic)
		refresh("writer", test.Questions[counter].Ask)
	}

	in = ""
	refresh("status", "Ready.")
	refresh("reader", "")
	resetFrame()
	return nil
}

func showTeam(v *gocui.View) error {
	in := response(v)
	refresh("reader", "")

	if isAnswer(in) { // correct answers show both: answer and congrats message
		player.Team = in
		player.Lead = lookUp(in)
		player.Score = completed()
		logTeam()
		updatePlayer()
		ymlFile = "0"
		resetTerm(ymlFile)
	}
	return nil
}

func logTeam() {

	message := []string{string(time.Now().Format(time.RFC822)), player.Name, player.Team, ymlFile, test.Topic, "TEAM UPDATE"}
	csvFile.Write(message)
	csvFile.Flush()

	message = []string{string(time.Now().Format(time.RFC822)), player.Name, player.Team, player.Lead, strconv.Itoa(player.Score), "", ""}
	csvStats.Write(message)
	csvStats.Flush()
}

func oldLookUp(s string, n string) string {

	var errBytes bytes.Buffer
	lookup := shLkup + " " + s + " " + n + " " + adminF
	cmd := exec.Command("/usr/bin/bash", "--noprofile", "-c", lookup)
	stdin, err := cmd.StdinPipe()
	EOK(errDir, err, "couldn't create input pipe", errBytes.String())
	defer stdin.Close()
	out, err := cmd.CombinedOutput()
	EOK(errDir, err, "couldn't start stdin", errBytes.String())
	return strings.TrimSuffix(string(out), "\n")
}

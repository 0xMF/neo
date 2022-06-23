package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/awesome-gocui/gocui"
)

type size struct {
	name   string
	top    int
	left   int
	bottom int
	right  int
}

func ok(err error, s ...string) error {
	if err != nil {
		switch len(s) {
		case 0:
			return err
		default:
			fmt.Print(s[0])
			return errors.Unwrap(fmt.Errorf("%w: %s", err, s[0]))
		}
	}
	return nil
}

func OK(err error, s ...string) {
	if err != nil {
		switch len(s) {
		case 0:
			log.Fatal(err)
		default:
			log.Fatal(s[0])
			log.Fatal(errors.Unwrap(fmt.Errorf("%w: %s", err, s[0])))
		}
	}
}

func EOK(epath string, err error, s ...string) {
	if err != nil {
		var elog = err

		// month as abbrev 3 letter and month as abbrev two digit integer
		//	t := time.Now()
		//	d := fmt.Sprintf("%d-%03s-%02d--%02d:%02d:%02d", t.Year(), t.Month().String()[:3], t.Day(), t.Hour(), t.Minute(), t.Second())
		//	d := fmt.Sprintf("%d-%02d-%02dT%02d:%02d:%02d", t.Year(), int(t.Month()), t.Day(), t.Hour(), t.Minute(), t.Second())
		//	e += "/" + d + "_" + usrname

		e, _ := filepath.EvalSymlinks(epath)
		e += "/" + usrname
		w := ""

		for _, r := range s {
			w += fmt.Sprintf("%s\n", r)
		}
		file, err := os.OpenFile(e, os.O_APPEND|os.O_CREATE|os.O_WRONLY, logRWrite)
		defer file.Close()
		err = os.Chmod(e, logRWrite)
		OK(err)

		// RFC822  = "02 Jan 06 15:04 MST"
		// RFC3339 = "2006-01-02T15:04:05Z07:00"
		//ioutil.WriteFile(e, []byte(fmt.Sprintf("%s: %v\n%s", time.Now().Format(time.RFC822), elog,w)),0240)
		file.WriteString(fmt.Sprintf("%s: %v\n%s", time.Now().Format(time.RFC822), elog, w))
		log.Printf("ERROR: %s\n\n", s[0])
		log.Printf("Email: %s\n", replyTo)
		log.Print("       with a screenshot/picture of this message and a")
		log.Fatal("       description of what happened before this error.")
	}
}

func layout(g *gocui.Gui) error {
	var sizes map[string]size = make(map[string]size)
	sizes["header"] = size{name: "header", top: -1, left: -1, bottom: term.width, right: 1}
	sizes["writer"] = size{name: "writer", top: -1, left: 1, bottom: term.width, right: term.height - 4}
	sizes["reader"] = size{name: "reader", top: -1, left: term.height - 4, bottom: term.width, right: term.height - 2}
	sizes["status"] = size{name: "status", top: -1, left: term.height - 2, bottom: term.width, right: term.height}
	//sizes["header"] = size{name: "header", top: 0, left: -1, bottom: term.width -1, right: 1}
	//sizes["writer"] = size{name: "writer", top: 0, left: 1, bottom: term.width -1, right: term.height - 4}
	//sizes["reader"] = size{name: "reader", top: 0, left: term.height - 5, bottom: term.width -1, right: term.height - 3}
	//sizes["status"] = size{name: "status", top: 0, left: term.height - 3, bottom: term.width -1, right: term.height - 1}

	for name, details := range sizes {
		var h handle = term.views[name]
		if h.View == nil {
			v, err := makeView(g, details)
			OK(err)
			term.views[name] = handle{View: v, text: h.text}
			refresh(name, h.text)
		}
	}

	if _, err := g.SetCurrentView("reader"); err != nil {
		return err
	}

	return nil
}

func cursorUp(g *gocui.Gui, v *gocui.View) error {
	if v != nil {
		ox, oy := v.Origin()
		cx, cy := v.Cursor()
		if err := v.SetCursor(cx, cy-1); err != nil && oy > 0 {
			if err := v.SetOrigin(ox, oy-1); err != nil {
				return err
			}
		}
	}
	return nil
}

func cursorDown(g *gocui.Gui, v *gocui.View) error {
	if v != nil {
		cx, cy := v.Cursor()
		if err := v.SetCursor(cx, cy+1); err != nil {
			ox, oy := v.Origin()
			if err := v.SetOrigin(ox, oy+1); err != nil {
				return err
			}
		}
	}
	return nil
}

func delMsg(g *gocui.Gui, v *gocui.View) error {
	if err := g.DeleteView("msg"); err != nil {
		return err
	}
	if _, err := g.SetCurrentView("reader"); err != nil {
		return err
	}
	return nil
}

func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}

func keybindings(g *gocui.Gui) error {
	if err := g.SetKeybinding("", gocui.KeyCtrlD, gocui.ModNone, quit); err != nil {
		return err
	}

	if err := g.SetKeybinding("reader", gocui.KeyTab, gocui.ModNone, nextView); err != nil {
		return err
	}
	if err := g.SetKeybinding("writer", gocui.KeyTab, gocui.ModNone, nextView); err != nil {
		return err
	}
	if err := g.SetKeybinding("writer", gocui.KeyArrowDown, gocui.ModNone, cursorDown); err != nil {
		return err
	}
	if err := g.SetKeybinding("writer", gocui.KeyArrowUp, gocui.ModNone, cursorUp); err != nil {
		return err
	}
	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		return err
	}
	if err := g.SetKeybinding("reader", gocui.KeyEnter, gocui.ModNone, check); err != nil {
		return err
	}
	if err := g.SetKeybinding("msg", gocui.KeyEnter, gocui.ModNone, delMsg); err != nil {
		return err
	}
	return nil
}

func makeView(g *gocui.Gui, s size) (*gocui.View, error) {
	v, err := g.SetView(s.name, s.top, s.left, s.bottom, s.right, 1)
	if err != gocui.ErrUnknownView {
		return nil, err
	}

	v.Frame = false
	v.FrameColor = gocui.ColorCyan
	v.Wrap = true

	if s.name == "reader" {
		v.Editable = true
	} else {
		v.Editable = false
	}

	if s.name == "header" || s.name == "status" {
		v.Highlight = true
	}

	return v, nil
}

func nextView(g *gocui.Gui, v *gocui.View) error {
	if v == nil || v.Name() == "reader" {
		_, err := g.SetCurrentView("writer")
		return err
	}
	_, err := g.SetCurrentView("reader")
	return err
}

// one refresh to rule them all
func refresh(n string, s string) {
	v := term.views[n].View
	v.Clear()
	v.SetCursor(0, 0)
	v.Frame = true
	term.Update(func(g *gocui.Gui) error {
		v, err := g.View(n)
		if err != nil {
			// handle error
		}
		v.Clear()
		fmt.Fprintf(v, "%v", s)
		return nil
	})
	term.views[n] = handle{View: v, text: s}
}

func header(v *gocui.View) error {
	fmt.Fprintf(v, "\n\t(%v/%v) %v\n\n", counter+1, len(test.Questions), test.Topic)
	v.Editable = true
	v.Highlight = true
	return nil
}

func writer(v *gocui.View) error {
	fmt.Fprintf(v, "\n\n%v\n\n", test.Questions[counter].Ask)
	v.Editable = true
	v.Wrap = true
	return nil
}

func reader(v *gocui.View) error {
	v.Editable = true
	v.Wrap = true
	fmt.Fprintf(v, term.views["reader"].text)
	return nil
}

func status(v *gocui.View) error {
	v.Editable = true
	v.Highlight = true
	fmt.Fprintf(v, term.views["status"].text)
	return nil
}

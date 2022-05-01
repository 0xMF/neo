package main

import (
	"errors"
	"fmt"
	"log"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/jroimartin/gocui"
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
			log.Fatal(errors.Unwrap(fmt.Errorf("%w: %s", err, s[0])))
		}
	}
}

func EOK(epath string, err error, s ...string) {
	if err != nil {
		t := time.Now()
		// month as abbrev 3 letter
		//d := fmt.Sprintf("%d-%03s-%02d--%02d:%02d:%02d", t.Year(), t.Month().String()[:3], t.Day(), t.Hour(), t.Minute(), t.Second())

		// month as abbrev two digit integer
		d := fmt.Sprintf("%d-%02d-%02d--%02d:%02d:%02d", t.Year(), int(t.Month()), t.Day(), t.Hour(), t.Minute(), t.Second())
		e, _ := filepath.EvalSymlinks(epath)
		e += "/" + d + "--" + usrname
		w := ""

		for _,r  := range s {
			w += fmt.Sprintf("%s\n",r)
		}
		file, err := os.OpenFile(e, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0240)
		OK(err)
		err = os.Chmod(e, 0244)
		OK(err)
		defer file.Close()

		ioutil.WriteFile(e, []byte(fmt.Sprintf("%s: %v\n%s", time.Now().Format(time.RFC822), err,w)),0240)
		log.Printf("%s\n", w)
		log.Print("ERROR: email " + sendTo)
		log.Print("       with a screenshot/picture of this error and a description")
		log.Fatal("       of what portion of the labs you were in before the error.")
	}
}

func layout(g *gocui.Gui) error {
	var sizes map[string]size = make(map[string]size)
	sizes["header"] = size{name: "header", top: -1, left: -1, bottom: term.width, right: 1}
	sizes["writer"] = size{name: "writer", top: -1, left: 1, bottom: term.width, right: term.height - 4}
	sizes["reader"] = size{name: "reader", top: -1, left: term.height - 4, bottom: term.width, right: term.height - 2}
	sizes["status"] = size{name: "status", top: -1, left: term.height - 2, bottom: term.width, right: term.height}

	for name, details := range sizes {
		var h handle = term.views[name]
		if h.View == nil {
			v, err := makeView(g, details)
			ok(err)
			term.views[name] = handle{View: v, call: h.call, text: h.text}
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
	v, err := g.SetView(s.name, s.top, s.left, s.bottom, s.right)
	if err != gocui.ErrUnknownView {
		return v, err
	}
	var h handle = term.views[s.name]
	ok(h.call(v), "cannot update "+s.name)
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
func refresh(n string, g *gocui.Gui, s string) error {
	var sizes map[string]size = make(map[string]size)
	ok(g.DeleteView(n))
	term.views[n] = handle{View: nil, call: term.views[n].call, text: s}
	v, err := makeView(g, sizes[n])
	term.views[n] = handle{View: v, call: term.views[n].call, text: s}
	return err
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

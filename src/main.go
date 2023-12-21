package main

import (
	"fmt"
	
	"github.com/jroimartin/gocui"
)

var OUTPUT_VIEW *gocui.View

func output_print(message string) {
	OUTPUT_VIEW.Clear()
	fmt.Fprint(OUTPUT_VIEW, message)
}

func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}

func for_mouse(g *gocui.Gui, v *gocui.View) error {
	cx, cy := v.Cursor()
	if _, err := v.Line(cy); err != nil {
	} else {
		check(v, cx, cy)
	}

	return nil
}

func not_editable(g *gocui.Gui, v *gocui.View) error {
	v.Editable = false

	return nil
}

func main() {
	//fmt.Println(VIEW_SET["main"][0].output())

	g, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
	}

	defer g.Close()

	if err = g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		//return err
	}

	if err = g.SetKeybinding("", gocui.KeyEnter, gocui.ModNone, not_editable); err != nil {
		//return err
	}

	//winX, winY := g.Size()

	g.Cursor = true
	g.Mouse = true

	if v, err := g.SetView("main", 0, 0, 50, 20); err != nil {
		if err != gocui.ErrUnknownView {
			return
		}
		v.Frame = true
		v.Title = "tlgGoAPI cli"

		view_set_init(v)
		build_view(v)

		if _, err := g.SetCurrentView("main"); err != nil {
			return
		}
	}

	if v, err := g.SetView("output", 0, 23, 80, 23+10); err != nil {
		if err != gocui.ErrUnknownView {
			return
		}

		OUTPUT_VIEW = v

		v.Frame = true
		v.Title = "Output"
		v.Wrap = true
	}

	if err = g.SetKeybinding("main", gocui.MouseRelease, gocui.ModNone, for_mouse); err != nil {
	}

	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
	}
}

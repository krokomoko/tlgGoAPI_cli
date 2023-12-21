package main

import (
	"fmt"
	"strings"

	"github.com/jroimartin/gocui"
)

const (
	ColorDefault = iota
	ColorBlack
	ColorRed
	ColorGreen
	ColorYellow
	ColorBlue
	ColorPurple
	ColorCyan
	ColorWhite
)

var COLOR_MAP = map[int]string{
	ColorDefault: "\033[0;30m",
	ColorBlack:   "\033[0;30m",
	ColorRed:     "\033[0;31m",
	ColorGreen:   "\033[0;32m",
	ColorYellow:  "\033[0;33m",
	ColorBlue:    "\033[0;34m",
	ColorPurple:  "\033[0;35m",
	ColorCyan:    "\033[0;36m",
	ColorWhite:   "\033[0;37m",
}

var VIEW_BUFFER = []string{"main"}

var VIEW_SET map[string][]GuiElement

var SIMPLE_CALLBACK, NEW_MENU bool

var ADD_BUTTON_CURRENT_IND = []int{4}

var NEW_WINDOW bool

var NEW_CALLBACK_NAME, NEW_CALLBACK_FROM_STATE, NEW_CALLBACK_CONTENT string

type GuiElement interface {
	output() string
	coords() (int, int, int, int)
	get_callback() func(v *gocui.View, x, y int)
	me(x, y int) bool
}

type Text struct {
	x        int
	y        int
	content  string
	color    int
	callback func(v *gocui.View, x, y int)
}

type Button struct {
	x        int
	y        int
	content  string
	color    int
	callback func(v *gocui.View, x, y int)
}

type Input struct {
	x        int
	y        int
	label    string
	content  string
	color    int
	callback func(v *gocui.View, x, y int)
}

type Checkbox struct {
	x        int
	y        int
	content  string
	color    int
	checked  bool
	callback func(v *gocui.View, x, y int)
}

func set_view(view_name string) {
	if view_name == "main" {
		VIEW_BUFFER = []string{"main"}
	} else {
		VIEW_BUFFER = append(VIEW_BUFFER, view_name)
	}
}

func cur_view() string {
	return VIEW_BUFFER[len(VIEW_BUFFER)-1]
}

func (t *Text) output() string {
	return fmt.Sprintf("%s%s\033[0m", COLOR_MAP[t.color], t.content)
}
func (t *Text) coords() (int, int, int, int) {
	return t.x, t.y, len(t.content), 0
}
func (t *Text) get_callback() func(v *gocui.View, x, y int) {
	return t.callback
}
func (t *Text) me(x, y int) bool {
	if y == t.y && x >= t.x && x <= t.x+len(t.content) {
		return true
	}
	return false
}

func (b *Button) output() string {
	return fmt.Sprintf(
		"%s\033[107m%s\033[0m",
		COLOR_MAP[b.color],
		b.content,
	)
}
func (b *Button) coords() (int, int, int, int) {
	return b.x, b.y, len(b.content), 1
}
func (b *Button) get_callback() func(v *gocui.View, x, y int) {
	return b.callback
}
func (b *Button) me(x, y int) bool {
	if y == b.y && x >= b.x && x <= b.x+len(b.content) {
		return true
	}
	return false
}

func (i *Input) output() string {
	return fmt.Sprintf("%s%s\033[0m:\n%s> %s", COLOR_MAP[ColorYellow], i.label, strings.Repeat(" ", i.x), i.content)
}
func (i *Input) coords() (int, int, int, int) {
	return i.x, i.y, 2 + len(i.content), 2
}
func (i *Input) get_callback() func(v *gocui.View, x, y int) {
	return i.callback
}
func (i *Input) me(x, y int) bool {
	if y == i.y+1 {
		return true
	}
	return false
}

func (ch *Checkbox) output() string {
	var checkbox_symbol = "\u2610"
	if ch.checked {
		checkbox_symbol = "\u2705"
	}

	return fmt.Sprintf(
		"%s%s %s%s\033[0m",
		checkbox_symbol,
		COLOR_MAP[ch.color],
		ch.content,
	)
}
func (ch *Checkbox) coords() (int, int, int, int) {
	var checkbox_symbol = "\u2610"
	if ch.checked {
		checkbox_symbol = "\u2705"
	}

	return ch.x, ch.y, len(fmt.Sprintf(
		"%s %s",
		checkbox_symbol,
		ch.content,
	)), 1
}
func (ch *Checkbox) get_callback() func(v *gocui.View, x, y int) {
	return ch.callback
}
func (ch *Checkbox) me(x, y int) bool {
	if y == ch.y && x >= ch.x && x <= ch.x+len(ch.content) {
		return true
	}
	return false
}

func get_input(v *gocui.View, input GuiElement) string {
	_, y, _, _ := input.coords()
	line_content, _ := v.Line(y + 1)
	return strings.TrimSpace(strings.Replace(line_content, "> ", "", 1))
}

func view_set_init(v *gocui.View) {
	VIEW_SET = make(map[string][]GuiElement)

	width, height := v.Size()

	VIEW_SET["main"] = []GuiElement{
		&Button{
			x:        width - 8,
			y:        0,
			content:  "Reset",
			color:    ColorRed,
			callback: reset,
		},
		&Button{
			x:        4,
			y:        1,
			content:  "> Add callback",
			color:    ColorBlue,
			callback: add_callback_button,
		},
		&Button{
			x:        4,
			y:        2,
			content:  "> Add menu",
			color:    ColorBlue,
			callback: add_menu,
		},
		&Button{
			x:        4,
			y:        3,
			content:  "> Change menu",
			color:    ColorBlue,
			callback: menu_list,
		},
		&Input{
			x:        4,
			y:        5,
			label:    "Work directory",
			color:    ColorYellow,
			callback: label,
		},
		&Button{
			x:       4,
			y:       height - 1,
			content: "Save",
			color:   ColorGreen,
			callback: save,
		},
	}

	VIEW_SET["add_callback"] = []GuiElement{
		&Button{
			x:        width - 8,
			y:        0,
			content:  "Reset",
			color:    ColorRed,
			callback: reset,
		},
		&Button{
			x:        4,
			y:        1,
			content:  "> Get user input",
			color:    ColorBlue,
			callback: get_user_input,
		},
		&Button{
			x:        4,
			y:        2,
			content:  "> Jump to menu",
			color:    ColorBlue,
			callback: menu_list,
		},
		&Button{
			x:        width - 7,
			y:        height - 1,
			content:  "Back",
			color:    ColorPurple,
			callback: back,
		},
	}

	VIEW_SET["from_state"] = []GuiElement{
		&Button{
			x:        width - 8,
			y:        0,
			content:  "Reset",
			color:    ColorRed,
			callback: reset,
		},
		&Input{
			x:        4,
			y:        1,
			label:    "Callback name",
			color:    ColorYellow,
			callback: label,
		},
		&Input{
			x:        4,
			y:        4,
			label:    "From state",
			color:    ColorYellow,
			callback: label,
		},
		&Input{
			x:        4,
			y:        7,
			label:    "Content",
			color:    ColorYellow,
			callback: label,
		},
		&Button{
			x:        4,
			y:        height - 1,
			content:  "Next",
			color:    ColorGreen,
			callback: next,
		},
		&Button{
			x:        width - 7,
			y:        height - 1,
			content:  "Back",
			color:    ColorPurple,
			callback: back,
		},
	}

	VIEW_SET["user_inputs"] = []GuiElement{
		&Button{
			x:        width - 8,
			y:        0,
			content:  "Reset",
			color:    ColorRed,
			callback: reset,
		},
		&Button{
			x:        4,
			y:        1,
			content:  "> Only text",
			color:    ColorBlue,
			callback: set_callback_type,
		},
		&Button{
			x:        4,
			y:        2,
			content:  "> Command",
			color:    ColorBlue,
			callback: set_callback_type,
		},
		&Button{
			x:        4,
			y:        3,
			content:  "> URL",
			color:    ColorBlue,
			callback: set_callback_type,
		},
		&Button{
			x:        4,
			y:        4,
			content:  "> Image",
			color:    ColorBlue,
			callback: set_callback_type,
		},
		&Button{
			x:        4,
			y:        5,
			content:  "> GIF",
			color:    ColorBlue,
			callback: set_callback_type,
		},
		&Button{
			x:        4,
			y:        6,
			content:  "> Sticker",
			color:    ColorBlue,
			callback: set_callback_type,
		},
		&Button{
			x:        4,
			y:        7,
			content:  "> Video",
			color:    ColorBlue,
			callback: set_callback_type,
		},
		&Button{
			x:        width - 7,
			y:        height - 1,
			content:  "Back",
			color:    ColorPurple,
			callback: back,
		},
	}

	VIEW_SET["jump_to_menu"] = []GuiElement{
		&Button{
			x:        width - 8,
			y:        0,
			content:  "Reset",
			color:    ColorRed,
			callback: reset,
		},
		&Button{
			x:        width - 7,
			y:        height - 1,
			content:  "Back",
			color:    ColorPurple,
			callback: back,
		},
	}

	VIEW_SET["add_menu"] = []GuiElement{
		&Button{
			x:        width - 8,
			y:        0,
			content:  "Reset",
			color:    ColorRed,
			callback: reset,
		},
		&Input{
			x:        4,
			y:        1,
			label:    "Menu name",
			color:    ColorYellow,
			callback: label,
		},
		&Input{
			x:        4,
			y:        4,
			label:    "From state",
			color:    ColorYellow,
			callback: label,
		},
		&Button{
			x:        4,
			y:        height - 1,
			content:  "Next",
			color:    ColorGreen,
			callback: next,
		},
		&Button{
			x:        width - 7,
			y:        height - 1,
			content:  "Back",
			color:    ColorPurple,
			callback: back,
		},
	}

	VIEW_SET["window_settings"] = []GuiElement{
		&Button{
			x:        width - 8,
			y:        0,
			content:  "Reset",
			color:    ColorRed,
			callback: reset,
		},
		&Input{
			x:        4,
			y:        1,
			label:    "Window name",
			color:    ColorYellow,
			callback: label,
		},
		&Input{
			x:        4,
			y:        4,
			label:    "Content",
			color:    ColorYellow,
			callback: label,
		},
		&Button{
			x:        4,
			y:        7,
			content:  "...",
			color:    ColorBlue,
			callback: add_button,
		},
		&Button{
			x:        4,
			y:        height - 1,
			content:  "Next",
			color:    ColorGreen,
			callback: menu_settings,
		},
		&Button{
			x:        width - 7,
			y:        height - 1,
			content:  "Back",
			color:    ColorPurple,
			callback: back,
		},
	}

	VIEW_SET["menu_settings"] = []GuiElement{
		&Button{
			x:        width - 8,
			y:        0,
			content:  "Reset",
			color:    ColorRed,
			callback: reset,
		},
		&Input{
			x:        4,
			y:        1,
			label:    "Menu name",
			color:    ColorYellow,
			callback: label,
		},
		&Input{
			x:        4,
			y:        4,
			label:    "From state",
			color:    ColorYellow,
			callback: label,
		},
		&Button{
			x:        4,
			y:        7,
			content:  "Change window",
			color:    ColorBlue,
			callback: open_window_list,
		},
		&Button{
			x:        4,
			y:        8,
			content:  "Add window",
			color:    ColorBlue,
			callback: add_new_window,
		},
		&Button{
			x:       4,
			y:       9,
			content: "View menu",
			color:   ColorBlue,
			//callback: view_menu,
		},
		&Button{
			x:       4,
			y:       height - 1,
			content: "Delete",
			color:   ColorRed,
			//callback: delete_menu,
		},
		&Button{
			x:        width - 7,
			y:        height - 1,
			content:  "Back",
			color:    ColorPurple,
			callback: back,
		},
	}

	VIEW_SET["set_button"] = []GuiElement{
		&Button{
			x:        width - 8,
			y:        0,
			content:  "Reset",
			color:    ColorRed,
			callback: reset,
		},
		&Input{
			x:        4,
			y:        1,
			label:    "Button name",
			color:    ColorYellow,
			callback: label,
		},
		&Input{
			x:        4,
			y:        4,
			label:    "Button text",
			color:    ColorYellow,
			callback: label,
		},
		&Button{
			x:        width - 7,
			y:        height - 3,
			content:  "Delete",
			color:    ColorRed,
			callback: delete_button,
		},
		&Button{
			x:        4,
			y:        height - 1,
			content:  "Next",
			color:    ColorGreen,
			callback: next,
		},
		&Button{
			x:        18,
			y:        height - 1,
			content:  "Set callback",
			color:    ColorBlue,
			callback: add_callback_button,
		},
		&Button{
			x:        width - 7,
			y:        height - 1,
			content:  "Back",
			color:    ColorPurple,
			callback: back,
		},
	}

	VIEW_SET["menu_list"] = []GuiElement{}

	VIEW_SET["window_list"] = []GuiElement{}
}

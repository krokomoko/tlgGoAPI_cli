package main

import (
	"fmt"

	"github.com/jroimartin/gocui"
)

func build_view(v *gocui.View) {
	v.Clear()

	var current_x, current_y int = 0, 0

	for _, el := range VIEW_SET[cur_view()] {
		x, y, width, height := el.coords()

		var dy int = y - current_y

		for i := 0; i < dy; i++ {
			fmt.Fprintln(v, " ")
		}

		if dy > 0 {
			current_x = 0
		}
		var dx int = x - current_x

		for i := 0; i < dx; i++ {
			fmt.Fprint(v, " ")
		}

		fmt.Fprint(v, el.output())

		current_x = x + width

		current_y = y + height - 1
	}
}

func check(v *gocui.View, x, y int) {
	v.Editable = false

	for _, el := range VIEW_SET[cur_view()] {
		if el.me(x, y) {
			callback := el.get_callback()
			if callback != nil {
				callback(v, x, y)
			}
		}
	}
}

func save(v *gocui.View, x, y int) {
	DIRECTORY = get_input(v, VIEW_SET["main"][4])
	if string(DIRECTORY[len(DIRECTORY)-1]) != "/" {
		DIRECTORY += "/"
	}

	if 0 == len(DIRECTORY) {
		output_print("Need to enter work directory before save")
	}

	err := tlg_ui_save()
	if err != nil {
		output_print(fmt.Sprintf("%s", err))
	} else {
		output_print("All saved")
	}
}

func back(v *gocui.View, x, y int) {
	v.Title = "Back..."

	if VIEW_BUFFER[len(VIEW_BUFFER)-1] == "set_button" {
		CURRENT_BUTTON = nil
	}

	if VIEW_BUFFER[len(VIEW_BUFFER)-1] == "window_settings" {
		VIEW_BUFFER[len(VIEW_BUFFER)-2] = "menu_settings"
	}

	if VIEW_BUFFER[len(VIEW_BUFFER)-1] == "menu_settings" {
		CURRENT_MENU = nil
		CURRENT_WINDOW = nil
		CURRENT_BUTTON = nil

		set_view("main")
		build_view(v)
		return
	}

	VIEW_BUFFER = VIEW_BUFFER[:len(VIEW_BUFFER)-1]

	if VIEW_BUFFER[len(VIEW_BUFFER)-1] == "main" {
		CURRENT_MENU = nil
		CURRENT_WINDOW = nil
		CURRENT_BUTTON = nil
		CURRENT_WINDOW_BUTTON_ROW_IND = 0
		CURRENT_WINDOW_BUTTON_IND = 0
	}

	SIMPLE_CALLBACK = false
	NEW_MENU = false

	build_view(v)
}

func reset(v *gocui.View, x, y int) {
	build_view(v)
}

func label(v *gocui.View, x, y int) {
	v.Editable = true
	for _, el := range VIEW_SET[cur_view()] {
		if el.me(x, y) {
			x_, _, width, _ := el.coords()
			v.SetCursor(x_+width, y)
		}
	}
}

func add_callback_button(v *gocui.View, x, y int) {
	if CURRENT_BUTTON != nil {
		button_name := get_input(v, VIEW_SET["set_button"][1])
		button_text := get_input(v, VIEW_SET["set_button"][2])
		if 0 == len(button_text) || 0 == len(button_name) {
			return
		}

		CURRENT_BUTTON.name = button_name
		CURRENT_BUTTON.content = button_text

		NEW_CALLBACK_NAME = fmt.Sprintf(
			"button_input_%s_%s_%s_callback",
			CURRENT_MENU.name,
			CURRENT_WINDOW.name,
			CURRENT_BUTTON.name,
		)

		NEW_CALLBACK_FROM_STATE = fmt.Sprintf(
			"%s_%s_%s_callback",
			CURRENT_MENU.name,
			CURRENT_WINDOW.name,
			CURRENT_BUTTON.name,
		)
	}

	v.Title = "Add callback"

	set_view("add_callback")

	build_view(v)
}

func set_button_wrapper(v *gocui.View, x, y int) {
	if CURRENT_BUTTON != nil {
		VIEW_SET["set_button"][1] = &Input{
			x:        4,
			y:        1,
			label:    "Button name",
			color:    ColorYellow,
			callback: label,
			content:  CURRENT_BUTTON.name,
		}
		VIEW_SET["set_button"][2] = &Input{
			x:        4,
			y:        4,
			label:    "Button text",
			color:    ColorYellow,
			callback: label,
			content:  CURRENT_BUTTON.content,
		}
	}

	set_view("set_button")
}

func menu_list(v *gocui.View, x, y int) {
	width, height := v.Size()

	v.Title = "Menu list"

	VIEW_SET["menu_list"] = []GuiElement{
		&Button{
			x:        width - 8,
			y:        0,
			content:  "Reset",
			color:    ColorRed,
			callback: reset,
		},
	}

	for menu_ind_, menu := range MENUS {
		menu_ind := menu_ind_
		
		VIEW_SET["menu_list"] = append(
			VIEW_SET["menu_list"],
			&Button{
				x:       4,
				y:       1 + menu_ind,
				content: "> " + menu.name,
				color:   ColorBlue,
				callback: func(v *gocui.View, x, y int) {
					CURRENT_MENU = MENUS[menu_ind]
					if len(VIEW_BUFFER) == 2 && VIEW_BUFFER[0] == "main" {
						menu_settings(v, x, y)
					} else {
						window_list(MENUS[menu_ind], v, x, y)
						set_view("window_list")
						build_view(v)
					}
				},
			},
		)
	}

	VIEW_SET["menu_list"] = append(
		VIEW_SET["menu_list"],
		&Button{
			x:        width - 7,
			y:        height - 1,
			content:  "Back",
			color:    ColorPurple,
			callback: back,
		},
	)

	set_view("menu_list")

	build_view(v)
}

func window_list(menu *TlgMenu, v *gocui.View, x, y int) {
	width, height := v.Size()

	v.Title = fmt.Sprintf("Menu \"%s\" - windows list", menu.name)

	VIEW_SET["window_list"] = []GuiElement{
		&Button{
			x:        width - 8,
			y:        0,
			content:  "Reset",
			color:    ColorRed,
			callback: reset,
		},
	}

	for window_ind_, window := range menu.windows {
		window_ind := window_ind_
		
		VIEW_SET["window_list"] = append(
			VIEW_SET["window_list"],
			&Button{
				x:       4,
				y:       1 + window_ind,
				content: "> " + window.name,
				color:   ColorBlue,
				callback: func(v *gocui.View, x, y int) {
					if CURRENT_BUTTON != nil {
						CURRENT_BUTTON.to = menu.windows[window_ind]
						set_button_wrapper(v, x, y)
						build_view(v)
					} else if CURRENT_MENU != nil {
						CURRENT_WINDOW = menu.windows[window_ind]
						add_window(v, x, y)
					} else {
						set_view("main")
						build_view(v)
					}
				},
			},
		)
	}

	VIEW_SET["window_list"] = append(
		VIEW_SET["window_list"],
		&Button{
			x:        width - 7,
			y:        height - 1,
			content:  "Back",
			color:    ColorPurple,
			callback: back,
		},
	)
}

func open_window_list(v *gocui.View, x, y int) {
	window_list(CURRENT_MENU, v, x, y)
	set_view("window_list")
	build_view(v)
}

func set_callback_type(v *gocui.View, x, y int) {
	if 0 == len(NEW_CALLBACK_NAME) ||
		0 == len(NEW_CALLBACK_FROM_STATE) ||
		0 == len(NEW_CALLBACK_CONTENT) {
		return
	}

	var types [COUNT_OF_TYPES]bool

	switch y {
	case 1: // Only text
		types[1] = true
	case 2: // Command
		types[0] = true
	case 3: // URL
		types[9] = true
	case 4: // Image
		types[2] = true
	case 5: // GIF
		types[6] = true
	case 6: // Sticker
		types[8] = true
	case 7: // Video
		types[3] = true
	default:
		return
	}

	var tlg_callback = &CallbackRegisterRow{
		name:    NEW_CALLBACK_NAME,
		types:   types,
		state:   NEW_CALLBACK_FROM_STATE,
		content: NEW_CALLBACK_CONTENT,
	}

	tlg_callback_to_reg(tlg_callback)

	callback_to_reg(
		NEW_CALLBACK_NAME,
		NEW_CALLBACK_FROM_STATE,
		types,
	)

	if CURRENT_BUTTON != nil {
		CURRENT_BUTTON.callback = tlg_callback

		set_button_wrapper(v, x, y)
	} else {
		set_view("main")
	}

	build_view(v)
}

func get_user_input(v *gocui.View, x, y int) {
	v.Title = "Callback from state"

	VIEW_SET["from_state"][1] = &Input{
		x:        4,
		y:        1,
		label:    "Callback name",
		color:    ColorYellow,
		callback: label,
		content:  NEW_CALLBACK_NAME,
	}
	VIEW_SET["from_state"][2] = &Input{
		x:        4,
		y:        4,
		label:    "From state",
		color:    ColorYellow,
		callback: label,
		content:  NEW_CALLBACK_FROM_STATE,
	}

	set_view("from_state")

	build_view(v)
}

func user_inputs_from_main(v *gocui.View, x, y int) {
	// Need to get values

	set_view("user_inputs")

	build_view(v)
}

func user_inputs(v *gocui.View, x, y int) {
}

func jump_to_menu(v *gocui.View, x, y int) {
	v.Title = "Add callback - jump to menu"

	set_view("jump_to_menu")

	build_view(v)
}

func add_menu(v *gocui.View, x, y int) {
	v.Title = "Add new menu"

	NEW_MENU = true

	set_view("add_menu")

	build_view(v)
}

func window_settings_build(v *gocui.View) {
	width, height := v.Size()

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
			content:  CURRENT_WINDOW.name,
		},
		&Input{
			x:        4,
			y:        4,
			label:    "Content",
			color:    ColorYellow,
			callback: label,
			content:  CURRENT_WINDOW.content,
		},
		&Button{
			x:        4,
			y:        7,
			content:  "...",
			color:    ColorBlue,
			callback: add_button,
		},
	}

	var current_ind int = 4

	ADD_BUTTON_CURRENT_IND = []int{current_ind}

	for row_ind_, row := range CURRENT_WINDOW.keyboard {
		row_ind := row_ind_
		for button_ind_, button := range row {
			button_ind := button_ind_
			x_, _, width_, _ := VIEW_SET["window_settings"][len(VIEW_SET["window_settings"])-1].coords()
			VIEW_SET["window_settings"] = append(
				VIEW_SET["window_settings"],
				&Button{
					x:       x_ + width_ + 1,
					y:       7 + row_ind,
					content: button.name,
					color:   ColorBlue,
					callback: func(v *gocui.View, x, y int) {
						change_button(CURRENT_WINDOW.keyboard[row_ind][button_ind], v, x, y)
					},
				},
			)
			ADD_BUTTON_CURRENT_IND[row_ind]++
		}
		VIEW_SET["window_settings"] = append(
			VIEW_SET["window_settings"],
			&Button{
				x:        4,
				y:        7 + row_ind + 1,
				content:  "...",
				color:    ColorBlue,
				callback: add_button,
			},
		)
		ADD_BUTTON_CURRENT_IND = append(ADD_BUTTON_CURRENT_IND, ADD_BUTTON_CURRENT_IND[row_ind]+1)
	}

	VIEW_SET["window_settings"] = append(
		VIEW_SET["window_settings"],
		&Button{
			x:        4,
			y:        height - 1,
			content:  "Next",
			color:    ColorGreen,
			callback: next,
		},
	)
	VIEW_SET["window_settings"] = append(
		VIEW_SET["window_settings"],
		&Button{
			x:        width - 7,
			y:        height - 1,
			content:  "Back",
			color:    ColorPurple,
			callback: back,
		},
	)
}

func add_window(v *gocui.View, x, y int) {
	v.Title = "Window settings"

	set_view("window_settings")

	window_settings_build(v)

	build_view(v)
}

func add_new_window(v *gocui.View, x, y int) {
	var window = &TlgWindow{
		menu: CURRENT_MENU,
	}

	CURRENT_WINDOW = window

	NEW_WINDOW = true

	add_window(v, x, y)
}

func change_button(button *TlgButton, v *gocui.View, x, y int) {
	window_name := get_input(v, VIEW_SET["window_settings"][1])
	window_content := get_input(v, VIEW_SET["window_settings"][2])

	if 0 == len(window_name) || 0 == len(window_content) {
		return
	}

	CURRENT_WINDOW.name = window_name
	CURRENT_WINDOW.content = window_content

	CURRENT_BUTTON = button

	VIEW_SET["set_button"][1] = &Input{
		x:        4,
		y:        1,
		label:    "Button name",
		color:    ColorYellow,
		callback: label,
		content:  CURRENT_BUTTON.name,
	}
	VIEW_SET["set_button"][2] = &Input{
		x:        4,
		y:        4,
		label:    "Button text",
		color:    ColorYellow,
		callback: label,
		content:  CURRENT_BUTTON.content,
	}

	v.Title = "Set button"

	set_view("set_button")

	build_view(v)
}

func add_button(v *gocui.View, x, y int) {
	window_name := get_input(v, VIEW_SET["window_settings"][1])
	window_content := get_input(v, VIEW_SET["window_settings"][2])

	if 0 == len(window_name) || 0 == len(window_content) {
		return
	}

	CURRENT_WINDOW.name = window_name
	CURRENT_WINDOW.content = window_content

	var row int = y - 7

	if CURRENT_WINDOW.keyboard == nil {
		CURRENT_WINDOW.keyboard = [][]*TlgButton{
			[]*TlgButton{},
		}
	}

	CURRENT_WINDOW_BUTTON_ROW_IND = row

	if CURRENT_WINDOW_BUTTON_ROW_IND > len(CURRENT_WINDOW.keyboard)-1 {
		var new_row = []*TlgButton{}
		CURRENT_WINDOW.keyboard = append(CURRENT_WINDOW.keyboard, new_row)
	}

	CURRENT_WINDOW_BUTTON_IND = len(CURRENT_WINDOW.keyboard[CURRENT_WINDOW_BUTTON_ROW_IND])

	CURRENT_BUTTON = &TlgButton{
		signal: fmt.Sprintf(
			"%s|%s|%s|%d|%d",
			CURRENT_MENU.state,
			CURRENT_MENU.name,
			CURRENT_WINDOW.name,
			CURRENT_WINDOW_BUTTON_ROW_IND,
			CURRENT_WINDOW_BUTTON_IND,
		),
	}

	CURRENT_WINDOW.keyboard[CURRENT_WINDOW_BUTTON_ROW_IND] = append(
		CURRENT_WINDOW.keyboard[CURRENT_WINDOW_BUTTON_ROW_IND],
		CURRENT_BUTTON,
	)

	v.Title = "Set button"

	set_view("set_button")

	build_view(v)
}

func delete_button(v *gocui.View, x, y int) {
}

func menu_settings(v *gocui.View, x, y int) {
	v.Title = "Menu settings"

	NEW_MENU = false

	if CURRENT_MENU != nil {
		VIEW_SET["menu_settings"][1] = &Input{
			x:        4,
			y:        1,
			label:    "Menu name",
			color:    ColorYellow,
			callback: label,
			content:  CURRENT_MENU.name,
		}

		VIEW_SET["menu_settings"][2] = &Input{
			x:        4,
			y:        4,
			label:    "From state",
			color:    ColorYellow,
			callback: label,
			content:  CURRENT_MENU.state,
		}
	}

	set_view("menu_settings")

	build_view(v)
}

func next(v *gocui.View, x, y int) {
	switch cur_view() {
	case "from_state":
		NEW_CALLBACK_NAME = get_input(v, VIEW_SET["from_state"][1])
		NEW_CALLBACK_FROM_STATE = get_input(v, VIEW_SET["from_state"][2])
		NEW_CALLBACK_CONTENT = get_input(v, VIEW_SET["from_state"][3])

		if 0 == len(NEW_CALLBACK_NAME) || 0 == len(NEW_CALLBACK_FROM_STATE) || 0 == len(NEW_CALLBACK_CONTENT) {
			return
		}

		user_inputs_from_main(v, x, y)

	case "add_menu":
		menu_name := get_input(v, VIEW_SET["add_menu"][1])
		from_state := get_input(v, VIEW_SET["add_menu"][2])
		if 0 == len(menu_name) || 0 == len(from_state) {
			return
		}

		var menu = &TlgMenu{
			name:    menu_name,
			state:   from_state,
			windows: []*TlgWindow{},
		}
		MENUS = append(MENUS, menu)

		CURRENT_MENU = menu

		var window = &TlgWindow{
			menu: menu,
			name: "main",
		}

		CURRENT_WINDOW = window

		NEW_WINDOW = true

		add_window(v, x, y)

	case "set_button":
		button_name := get_input(v, VIEW_SET["set_button"][1])
		button_text := get_input(v, VIEW_SET["set_button"][2])
		if 0 == len(button_text) || 0 == len(button_name) {
			return
		}

		CURRENT_BUTTON.name = button_name
		CURRENT_BUTTON.content = button_text

		CURRENT_BUTTON = nil

		add_window(v, x, y)

	case "window_settings":
		window_name := get_input(v, VIEW_SET["window_settings"][1])
		window_content := get_input(v, VIEW_SET["window_settings"][2])

		if 0 == len(window_name) || 0 == len(window_content) {
			return
		}

		CURRENT_WINDOW.name = window_name
		CURRENT_WINDOW.content = window_content

		if NEW_WINDOW {
			NEW_WINDOW = false
			CURRENT_MENU.windows = append(CURRENT_MENU.windows, CURRENT_WINDOW)
		}

		CURRENT_WINDOW = nil
		CURRENT_BUTTON = nil

		menu_settings(v, x, y)
	}
}

package main

import (
	"fmt"
	"os"
)

const COUNT_OF_TYPES int = 11

var BUTTON_NAME string

type TlgButton struct {
	name     string
	content  string
	signal   string
	callback *CallbackRegisterRow
	to       *TlgWindow
}

type TlgWindow struct {
	menu     *TlgMenu
	name     string
	content  string
	keyboard [][]*TlgButton
}

type TlgMenu struct {
	name    string
	state   string
	windows []*TlgWindow
}

type CallbackRegisterRow struct {
	name    string
	types   [COUNT_OF_TYPES]bool
	state   string
	content string
}

var CALLBACK_REGISTER = []*CallbackRegisterRow{}
var TLG_CALLBACK_REGISTER = []*CallbackRegisterRow{}

var MESSAGE_TYPES_NAMES = map[int]string{
	0:  "M_COMMAND",
	1:  "M_TEXT",
	2:  "M_IMAGE",
	3:  "M_VIDEO",
	4:  "M_VOICE",
	5:  "M_AUDIO",
	6:  "M_GIF",
	7:  "M_DOCUMENT",
	8:  "M_STICKER",
	9:  "M_URL",
	10: "M_CALLBACK",
}

var MENUS []*TlgMenu

var CURRENT_MENU *TlgMenu
var CURRENT_WINDOW *TlgWindow
var CURRENT_BUTTON *TlgButton
var CURRENT_WINDOW_BUTTON_ROW_IND int
var CURRENT_WINDOW_BUTTON_IND int
var DIRECTORY string

func callback_to_reg(name, state string, types [COUNT_OF_TYPES]bool) {
	CALLBACK_REGISTER = append(
		CALLBACK_REGISTER,
		&CallbackRegisterRow{
			name:  name,
			types: types,
			state: state,
		},
	)
}

func tlg_callback_to_reg(tlg_callback *CallbackRegisterRow) {
	TLG_CALLBACK_REGISTER = append(
		TLG_CALLBACK_REGISTER,
		tlg_callback,
	)
}

func add_to_file(filename string, content string, commentary string) error {
	f, err := os.OpenFile(DIRECTORY+filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err = f.WriteString(fmt.Sprintf("\n/*\n%s*/\n%s\n", commentary, content)); err != nil {
		return err
	}

	return nil
}

func tlg_ui_save() error {
	tlg_ui_save_menu()

	tlg_ui_save_callback_functions()

	tlg_ui_save_callbacks()

	return nil
}

func tlg_ui_save_callback_functions() {
	var output string
	
	for _, row := range TLG_CALLBACK_REGISTER {
		output += fmt.Sprintf("func %s(bot *tlgGoAPI.Bot, update *tlgGoAPI.Update) error {\n	return nil\n}\n\n", row.name)
	}

	add_to_file("callbacks.go", output, "Callbacks")
}

func tlg_ui_save_callbacks() {
	var output string = "func bind_callbacks(bot *tlgGoAPI.Bot) {\n"
	var tab string = "	"

	for _, row := range CALLBACK_REGISTER {
		var message_types string = "[]uint8{"
		for type_ind, type_ := range row.types {
			if type_ {
				message_types += "tlgGoAPI." + MESSAGE_TYPES_NAMES[type_ind] + ", "
			}
		}
		if 0 == len(message_types) {
			//return "", errors.New("ERROR: - 0 message types set to" + row.name + " function")
		}
		message_types = message_types[:len(message_types)-2] + "}"

		output += tab + fmt.Sprintf(
			"bot.L(&%s, \"%s\", %s)\n",
			message_types,
			row.state,
			row.name,
		)
	}
	output += "}"

	add_to_file("main.go", output, "Callbacks bindings")
}

func compile_window_keyboard(window *TlgWindow) string {
	var result string = fmt.Sprintf("var %s_%s_keyboard = tlgGoAPI.InlineKeyboardMarkup{\n", window.menu.name, window.name)
	result += "	InlineKeyboard: [][]tlgGoAPI.InlineKeyboardButton{\n"
	for _, row := range (*window).keyboard {
		result += "		[]tlgGoAPI.InlineKeyboardButton{\n"
		for _, button := range row {
			result += "			tlgGoAPI.InlineKeyboardButton{\n"
			result += fmt.Sprintf("				Text: \"%s\",\n", button.content)
			result += fmt.Sprintf("				CallbackData: \"%s\",\n", button.signal)
			result += "			},\n"
		}
		result += "		},\n"
	}
	result += "	},\n"
	result += "}\n"

	return result
}

func compile_window_callback(window *TlgWindow) string {
	var callback_to_menu string

	// callbacks to get signal from window
	callback_to_menu = fmt.Sprintf("func %s_%s_callback(bot *tlgGoAPI.Bot, update *tlgGoAPI.Update) error {\n", window.menu.name, window.name)
	callback_to_menu += "	callback := update.CallbackQuery.Data\n"
	callback_to_menu += "	fromId := bot.GetFromId(update)\n"
	callback_to_menu += "	var upd tlgGoAPI.EditMessageText\n\n"
	callback_to_menu += "	switch callback {\n"
	for _, row := range window.keyboard {
		for _, button := range row {
			callback_to_menu += fmt.Sprintf("	case \"%s\":\n", button.signal)
			callback_to_menu += "		upd.ChatId = fromId\n"
			callback_to_menu += "		upd.MessageId = update.CallbackQuery.Message.MessageId\n"
			if button.to != nil {
				callback_to_menu += fmt.Sprintf("		upd.Text = `%s`\n", button.to.content)
				callback_to_menu += fmt.Sprintf("		upd.ReplyMarkup = %s_%s_keyboard\n", button.to.menu.name, button.to.name)
				callback_to_menu += fmt.Sprintf("		bot.SetUserState(fromId, \"%s_%s\")\n", button.to.menu.name, button.to.name)
			} else if button.callback != nil {
				callback_to_menu += fmt.Sprintf("		upd.Text = `%s`\n", button.callback.content)
				callback_to_menu += fmt.Sprintf("		bot.SetUserState(fromId, \"%s\")\n", button.callback.state)
			}
		}
	}
	callback_to_menu += "\n	}\n\n	bot.Call(upd)\n	return nil\n}"

	// bindings
	var types [COUNT_OF_TYPES]bool
	types[10] = true
	CALLBACK_REGISTER = append(
		CALLBACK_REGISTER,
		&CallbackRegisterRow{
			name:  fmt.Sprintf("%s_%s_callback", window.menu.name, window.name),
			types: types,
			state: fmt.Sprintf("%s_%s", window.menu.name, window.name),
		},
	)

	return callback_to_menu
}

func compile_menu(menu *TlgMenu) {
	var main_function_callback_name = fmt.Sprintf("%s_%s", menu.name, menu.state)

	var keyboards string
	for _, window := range menu.windows {
		keyboards += fmt.Sprintf("%s\n", compile_window_keyboard(window))
	}

	var callbacks string

	callbacks += fmt.Sprintf(
		"func %s_menu(bot *tlgGoAPI.Bot, update *tlgGoAPI.Update) error {\n",
		main_function_callback_name,
	)
	callbacks += "	fromId := bot.GetFromId(update)\n"
	callbacks += fmt.Sprintf("	bot.SetUserState(fromId, \"%s_%s\")\n", menu.name, menu.windows[0].name)
	callbacks += "	upd := tlgGoAPI.SendMessage{\n"
	callbacks += "		ChatId: fromId,\n"
	callbacks += fmt.Sprintf("		Text: `%s`,\n", menu.windows[0].content)
	callbacks += fmt.Sprintf("		ReplyMarkup: %s_%s_keyboard,\n	}\n", menu.name, menu.windows[0].name)
	callbacks += "	bot.Call(upd)\n	return nil\n}\n\n"
	var types [COUNT_OF_TYPES]bool
	types[0] = true
	for ind, window := range menu.windows {
		if 0 == ind {
			continue
		}
		CALLBACK_REGISTER = append(
			CALLBACK_REGISTER,
			&CallbackRegisterRow{
				name:  main_function_callback_name + "_menu",
				types: types,
				state: fmt.Sprintf("%s_%s", menu.name, window.name),
			},
		)
	}
	CALLBACK_REGISTER = append(
		CALLBACK_REGISTER,
		&CallbackRegisterRow{
			name:  main_function_callback_name + "_menu",
			types: types,
			state: menu.state,
		},
	)

	callbacks += compile_window_callback(menu.windows[0]) + "\n\n"

	for ind, window := range menu.windows {
		if 0 == ind {
			continue
		}
		callbacks += compile_window_callback(window) + "\n\n"
	}

	var filename string = fmt.Sprintf("%s_menu.go", menu.name)

	add_to_file(filename, keyboards, "Keyboards\n")

	add_to_file(filename, callbacks, "Menu "+menu.name)

}

func tlg_ui_save_menu() error {
	for _, menu := range MENUS {
		compile_menu(menu)
	}

	return nil
}
